import { writable, derived, get } from "svelte/store";
import { browser } from "$app/environment";
import { auth } from './auth';
import { api } from '$lib/api/client';
import { isOffline } from './offline';
import { llmStore } from './llm';
import type { Message } from '$lib/types';
import { secureStorage } from '$lib/utils/secureStorage';
import { API_URL } from '$lib/config';

interface WebSocketState {
    connected: boolean;
    connecting: boolean;
    error: string | null;
    messages: Message[];
    messageQueue: Message[];
    loading: boolean;
    pendingMessage: Message | null;
    lastMessageId: string | null;
    retryCount: number;
}

const MAX_RETRIES = 5;
const RETRY_DELAY = 2000; // 2 seconds

type MessageCallback = (message: any) => void;

let ws: WebSocket | null = null;
let sessionId: string | undefined;
let manualClose = false;
let messageCallback: MessageCallback | undefined;

function createWebSocketStore() {
    const { subscribe, set, update } = writable<WebSocketState>({
        connected: false,
        connecting: false,
        error: null,
        messages: [],
        messageQueue: [],
        loading: false,
        pendingMessage: null,
        lastMessageId: null,
        retryCount: 0
    });

    let reconnectTimeout: NodeJS.Timeout | null = null;
    let retryTimeout: number | null = null;

    const loadMessages = async (beforeId?: string) => {
        if (!sessionId) return;

        try {
            const newMessages = await api.getSessionMessages(sessionId, beforeId);
            if (newMessages.length > 0) {
                update(state => ({
                    ...state,
                    messages: beforeId ? [...state.messages, ...newMessages] : newMessages,
                    lastMessageId: newMessages[newMessages.length - 1].id
                }));
            }
        } catch (err) {
            console.error("Failed to load messages:", err);
            update(state => ({
                ...state,
                error: "Failed to load messages"
            }));
        }
    };

    const connect = async (sid: string) => {
        if (!browser || !sid) return;

        sessionId = sid;
        manualClose = false;

        update(state => ({ ...state, connecting: true, error: null }));

        const authState = get(auth);
        if (!authState.token) {
            update(state => ({ ...state, connecting: false, error: 'No authentication token available' }));
            return;
        }

        console.log(`sessionId: ${sessionId}`);
        console.log(`token: ${authState.token}`);
        const wsUrl = `${API_URL.replace(/^http/, 'ws')}/ws?session_id=${encodeURIComponent(sessionId)}&token=${encodeURIComponent(authState.token)}`;
        console.log('Connecting to WebSocket:', wsUrl);

        try {
            console.log(new WebSocket(wsUrl));
            ws = new WebSocket(wsUrl);

            let connectTimeout = setTimeout(() => {
                if (ws && ws.readyState === WebSocket.CONNECTING) {
                    ws.close();
                    update(state => ({
                        ...state,
                        error: 'WebSocket connection timeout',
                        connected: false,
                        connecting: false
                    }));
                }
            }, 5000);

            ws.onopen = () => {
                clearTimeout(connectTimeout);
                console.log('WebSocket connection established');
                update(state => ({
                    ...state,
                    connected: true,
                    connecting: false,
                    error: null,
                    retryCount: 0
                }));

                sendQueuedMessages();
            };

            ws.onclose = (event) => {
                console.log('WebSocket connection closed:', event);
                update(state => ({ ...state, connected: false, connecting: false }));

                if (!manualClose && sessionId) {
                    const currentState = get({ subscribe });
                    if (currentState.retryCount < MAX_RETRIES) {
                        const retryDelay = Math.min(1000 * Math.pow(2, currentState.retryCount), 10000);
                        console.log(`Attempting to reconnect in ${retryDelay}ms... (attempt ${currentState.retryCount + 1}/${MAX_RETRIES})`);
                        setTimeout(() => {
                            update(state => ({ ...state, retryCount: state.retryCount + 1 }));
                            connect(sessionId!);
                        }, retryDelay);
                    } else {
                        console.error('Max reconnection attempts reached');
                        update(state => ({
                            ...state,
                            error: 'Failed to connect after multiple attempts. Please refresh the page.'
                        }));
                    }
                }
            };

            ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                update(state => ({ ...state, error: 'WebSocket connection error' }));
            };

            ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    if (message && typeof message === 'object') {
                        messageCallback?.(message);

                        update(state => ({
                            ...state,
                            messages: [...state.messages, message],
                            loading: false,
                            pendingMessage: null
                        }));
                    }
                } catch (error) {
                    console.error('Failed to parse WebSocket message:', error);
                }
            };
        } catch (error) {
            console.error('Failed to create WebSocket connection:', error);
            update(state => ({
                ...state,
                connected: false,
                connecting: false,
                error: 'Failed to create WebSocket connection'
            }));
        }
    }

    function disconnect() {
        manualClose = true;
        if (ws) {
            ws.close();
            ws = null;
        }
        if (reconnectTimeout) {
            clearTimeout(reconnectTimeout);
            reconnectTimeout = null;
        }
        if (retryTimeout) {
            clearTimeout(retryTimeout);
            retryTimeout = null;
        }
        sessionId = undefined;
        update(state => ({
            ...state,
            connected: false,
            connecting: false,
            error: null,
            messages: [],
            messageQueue: [],
            loading: false,
            pendingMessage: null,
            lastMessageId: null,
            retryCount: 0
        }));
    }

    function sendMessage(message: Message) {
        if (!ws || ws.readyState !== WebSocket.OPEN) {
            update(state => ({
                ...state,
                messageQueue: [...state.messageQueue, message]
            }));
            return;
        }

        const messageWithModel = {
            ...message,
            model: get(llmStore).selectedModel?.id || 'default'
        };

        try {
            ws.send(JSON.stringify(messageWithModel));
            update(state => ({
                ...state,
                messages: [...state.messages, messageWithModel],
                loading: true,
                pendingMessage: messageWithModel
            }));
        } catch (error) {
            console.error('Failed to send message:', error);
            update(state => ({
                ...state,
                error: 'Failed to send message',
                messageQueue: [...state.messageQueue, messageWithModel]
            }));

            const currentState = get({ subscribe });
            if (currentState.retryCount < MAX_RETRIES) {
                if (retryTimeout) {
                    clearTimeout(retryTimeout);
                }
                retryTimeout = window.setTimeout(() => {
                    update(state => ({ ...state, retryCount: state.retryCount + 1 }));
                    sendMessage(message);
                }, RETRY_DELAY);
            }
        }
    }

    function sendQueuedMessages() {
        const state = get({ subscribe });
        if (state.messageQueue.length > 0) {
            state.messageQueue.forEach(message => {
                sendMessage(message);
            });
            update(state => ({ ...state, messageQueue: [] }));
        }
    }

    const loadMoreMessages = () => {
        const currentState = get({ subscribe });
        if (currentState.lastMessageId) {
            loadMessages(currentState.lastMessageId);
        }
    };

    // Subscribe to auth store to handle disconnect when user logs out
    auth.subscribe((state: { user: { id: string } | null }) => {
        if (!state.user) {
            disconnect();
        }
    });

    return {
        subscribe,
        connect,
        disconnect,
        sendMessage,
        loadMessages,
        loadMoreMessages,
        onMessage(callback: MessageCallback) {
            messageCallback = callback;
        }
    };
}

// Create and export the store
export const store = createWebSocketStore();

// Derived stores
export const isReconnecting = derived(store, $store => $store.connecting && !$store.connected);
export const messages = derived(store, $store => $store.messages);
export const lastMessage = derived(store, $store => $store.messages[$store.messages.length - 1]);
export const isConnected = derived(store, $store => $store.connected);
export const isLoading = derived(store, $store => $store.loading);
export const error = derived(store, $store => $store.error);
