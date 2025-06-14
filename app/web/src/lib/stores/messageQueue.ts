import { browser } from '$app/environment';
import { writable, get } from 'svelte/store';
import { isOffline } from './offline';
import type { Message } from '$lib/types';

interface QueuedMessage {
  id: string;
  sessionId: string;
  content: string;
  timestamp: number;
  retryCount: number;
}

const MAX_RETRIES = 3;
const STORAGE_KEY = 'offline_message_queue';

export const messageQueue = writable<QueuedMessage[]>([]);

// Initialize from localStorage if available
if (browser) {
    const savedQueue = localStorage.getItem(STORAGE_KEY);
    if (savedQueue) {
        try {
            const parsed = JSON.parse(savedQueue);
            messageQueue.set(parsed);
        } catch (e) {
            console.error('Failed to parse message queue from localStorage:', e);
        }
    }

    // Subscribe to changes and persist to localStorage
    messageQueue.subscribe(queue => {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(queue));
    });
}

// Add a message to the queue
export function queueMessage(message: Omit<QueuedMessage, 'timestamp' | 'retryCount'>) {
    const queue = get(messageQueue);
    messageQueue.set([
        ...queue,
        {
            ...message,
            timestamp: Date.now(),
            retryCount: 0
        }
    ]);
}

// Remove a message from the queue
export function removeFromQueue(messageId: string) {
    const queue = get(messageQueue);
    messageQueue.set(queue.filter(msg => msg.id !== messageId));
}

// Process the queue when back online
export async function processQueue(sendMessage: (message: Omit<Message, 'id' | 'createdAt'>) => Promise<void>) {
    const queue = get(messageQueue);
    const offline = get(isOffline);
    
    if (offline || queue.length === 0) return;

    for (const message of queue) {
        try {
            const userId = browser ? localStorage.getItem('userId') || '' : '';
            await sendMessage({
                sessionId: message.sessionId,
                content: message.content,
                type: 'message',
                userId,
                model: 'default'
            });
            removeFromQueue(message.id);
        } catch (error) {
            if (message.retryCount < MAX_RETRIES) {
                messageQueue.update(queue => 
                    queue.map(msg => 
                        msg.id === message.id 
                            ? { ...msg, retryCount: msg.retryCount + 1 }
                            : msg
                    )
                );
            } else {
                removeFromQueue(message.id);
            }
        }
    }
} 