import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import { notifications } from './notifications';
import { processQueue } from './messageQueue';

export const isOffline = writable(browser ? !navigator.onLine : false);

let wasOffline = browser ? !navigator.onLine : false;

function updateOnlineStatus() {
    if (!browser) return;
    
    const nowOffline = !navigator.onLine;
    isOffline.set(nowOffline);

    if (wasOffline && !nowOffline) {
        notifications.add({
            type: 'success',
            message: 'Connection restored. Syncing messages...',
            duration: 3000
        });
    } else if (!wasOffline && nowOffline) {
        notifications.add({
            type: 'warning',
            message: 'You are offline. Messages will be sent when you reconnect.',
            duration: 0
        });
    }

    wasOffline = nowOffline;
}

if (browser) {
    window.addEventListener('online', updateOnlineStatus);
    window.addEventListener('offline', updateOnlineStatus);
} 