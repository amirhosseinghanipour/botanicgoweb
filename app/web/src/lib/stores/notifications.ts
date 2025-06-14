import { writable } from 'svelte/store';

export interface Notification {
  id: string;
  type: 'success' | 'error' | 'info' | 'warning';
  message: string;
  duration?: number;
}

function createNotifications() {
  const { subscribe, update } = writable<Notification[]>([]);

  return {
    subscribe,
    add: (notification: Omit<Notification, 'id'>) => {
      const id = crypto.randomUUID();
      update(notifications => [...notifications, { ...notification, id }]);
      
      if (notification.duration !== 0) {
        setTimeout(() => {
          update(notifications => notifications.filter(n => n.id !== id));
        }, notification.duration || 5000);
      }
    },
    remove: (id: string) => {
      update(notifications => notifications.filter(n => n.id !== id));
    }
  };
}

export const notifications = createNotifications(); 