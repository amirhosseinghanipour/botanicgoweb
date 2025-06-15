// web/src/lib/stores/websocket.ts

import { writable, derived, get } from "svelte/store";
import { browser } from "$app/environment";
import { auth } from './auth';
import { api } from '$lib/api/client';
import type { Message } from '$lib/types';
import { API_URL } from '$lib/config';
import { v4 as uuidv4 } from 'uuid'; // New import for UUID generation

interface WebSocketState {
  connected: boolean;
  connecting: boolean;
  error: string | null;
  messages: Message[];
  retryCount: number;
}

const MAX_RETRIES = 5;

type MessageCallback = (message: Message) => void;

let ws: WebSocket | null = null;
let sessionId: string | undefined;
let manualClose = false;
let messageCallback: MessageCallback | undefined;
let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;

function createWebSocketStore() {
  const { subscribe, set, update } = writable<WebSocketState>({
    connected: false,
    connecting: false,
    error: null,
    messages: [],
    retryCount: 0
  });

  const loadMessages = async (sid?: string) => {
    const currentSessionId = sid || sessionId;
    if (!currentSessionId) return;

    try {
      const newMessages = await api.getSessionMessages(currentSessionId);
      update(state => ({
        ...state,
        messages: newMessages,
        error: null,
      }));
    } catch (err) {
      console.error("Failed to load messages:", err);
      update(state => ({
        ...state,
        error: "Failed to load messages",
      }));
    }
  };

  const connect = async (sid: string) => {
    if (!browser || (ws && ws.readyState === WebSocket.OPEN)) {
      console.log('WebSocket connect: Already connected or not in browser environment.');
      return;
    }

    sessionId = sid;
    manualClose = false;

    update(state => ({ ...state, connecting: true, error: null }));

    const authState = get(auth);
    if (!authState.token) {
      console.error('WebSocket connect: No authentication token available. Cannot connect.');
      update(state => ({ ...state, connecting: false, error: 'No authentication token available' }));
      return;
    }

    const wsUrl = `${API_URL.replace(/^http/, 'ws')}/ws?session_id=${encodeURIComponent(sessionId)}&token=${encodeURIComponent(authState.token)}`;
    console.log('Attempting to connect WebSocket to URL:', wsUrl);
    console.log('Using token (first 10 chars):', authState.token.substring(0, 10), '...');

    try {
      ws = new WebSocket(wsUrl);

      ws.onopen = () => {
        console.log('WebSocket connection established!');
        if (reconnectTimeout) clearTimeout(reconnectTimeout);
        update(state => ({
          ...state,
          connected: true,
          connecting: false,
          error: null,
          retryCount: 0
        }));
      };

      ws.onclose = (event) => {
        console.log('WebSocket connection closed:', event.code, event.reason);
        ws = null;
        update(state => ({ ...state, connected: false, connecting: false }));
        if (!manualClose) {
          reconnect();
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error occurred:', error);
        update(state => ({ ...state, error: 'WebSocket connection error' }));
      };

      ws.onmessage = (event) => {
        console.log('Raw WebSocket event data:', event.data);
        try {
          const message: Message = JSON.parse(event.data);
          console.log('Parsed WebSocket message:', message);

          if (messageCallback) {
            console.log('Calling messageCallback...');
            messageCallback(message);
            console.log('messageCallback finished.');
          }

          update(state => {
            if (state.messages.some(m => m.id === message.id)) {
              console.log('Message with ID already exists, skipping:', message.id);
              return state;
            }
            console.log('Adding new message to store:', message.id);
            return {
              ...state,
              messages: [...state.messages, message],
            };
          });

        } catch (error) {
          console.error('Failed to parse WebSocket message or process:', error);
        }
      };

    } catch (error) {
      console.error('Failed to initialize WebSocket connection:', error);
      update(state => ({
        ...state,
        connecting: false,
        error: 'Failed to create WebSocket connection'
      }));
      reconnect();
    }
  };

  const reconnect = () => {
    const { retryCount } = get({ subscribe });
    if (retryCount >= MAX_RETRIES) {
      update(state => ({ ...state, error: "Connection failed. Please refresh." }));
      return;
    }

    const delay = Math.pow(2, retryCount) * 1000;
    console.log(`Attempting to reconnect in ${delay / 1000}s...`);

    if (reconnectTimeout) clearTimeout(reconnectTimeout);
    reconnectTimeout = setTimeout(() => {
      update(state => ({ ...state, retryCount: state.retryCount + 1 }));
      if (sessionId) connect(sessionId);
    }, delay);
  };

  function disconnect() {
    manualClose = true;
    if (reconnectTimeout) clearTimeout(reconnectTimeout);
    if (ws) {
      ws.close();
      console.log('WebSocket manually disconnected.');
    }
    set({
      connected: false,
      connecting: false,
      error: null,
      messages: [],
      retryCount: 0
    });
  }

  function sendMessage(messageContent: string, currentSessionId: string) {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      console.error("WebSocket is not connected. Cannot send message.");
      update(state => ({ ...state, error: "Cannot send message: not connected." }));
      return;
    }

    const authState = get(auth);
    const userId = authState.user?.id;

    if (!userId) {
      console.error("User not authenticated. Cannot send message.");
      return;
    }

    let finalContent: string;
    // Defensive check: ensure content is always a string
    if (typeof messageContent !== 'string') {
      console.warn('sendMessage received non-string content; attempting to convert to string.', messageContent);
      finalContent = String(messageContent);
    } else {
      finalContent = messageContent;
    }

    const message = {
      id: uuidv4(), // Generate a unique ID for the message
      type: 'message',
      role: 'user',
      content: finalContent, // Now guaranteed to be a string
      sessionId: currentSessionId, // Ensure sessionId is passed from calling component
      userId: userId,
      createdAt: new Date().toISOString()
    };

    console.log('Sending WebSocket message (final content as string):', message);
    ws.send(JSON.stringify(message));
  }

  auth.subscribe((state) => {
    if (!state.isAuthenticated && !state.isLoading) {
      console.log('Auth state changed to unauthenticated, disconnecting WebSocket.');
      disconnect();
    }
  });

  return {
    subscribe,
    connect,
    disconnect,
    sendMessage,
    loadMessages,
    onMessage(callback: MessageCallback) {
      messageCallback = callback;
    }
  };
}

export const store = createWebSocketStore();
export const messages = derived(store, $store => $store.messages);
export const isConnected = derived(store, $store => $store.connected);
export const isLoading = derived(store, $store => $store.connecting);
export const isReconnecting = derived(store, $store => $store.connecting && !$store.connected);
export const error = derived(store, $store => $store.error);
