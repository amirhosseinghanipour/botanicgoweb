import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';
import type { Message } from '$lib/types';

// Types
export interface ChatSession {
  id: string;
  user_id: string;
  title: string;
  model: string;
  created_at: string;
  updated_at: string;
  messages: Message[];
}

export interface Model {
  id: string;
  name: string;
  description: string;
  contextLength: number;
  pricing: {
    prompt: string;
    completion: string;
  };
}

// Theme
export const isDarkMode = writable(
  browser && localStorage.getItem("theme") === "dark"
);

export type ThemePreference = 'light' | 'dark' | 'system';
export const themePreference = writable<ThemePreference>(
  (browser && (localStorage.getItem("themePreference") as ThemePreference)) || 'system'
);

// Chat sessions
export const sessions = writable<ChatSession[]>([]);
export const activeSessionId = writable<string | null>(
  browser && localStorage.getItem("activeSessionId") || null
);

// Models
export const models = writable<Model[]>([]);

// Active model
export const activeModel = writable<string | null>(
  browser && localStorage.getItem("activeModel") || null
);

// Loading states
export const isLoading = writable(false);
export const error = writable<string | null>(null);
export const isTextareaFocused = writable(false);

// Prompt text
export const promptText = writable('');

// Selected model
export const selectedModelId = writable<string | null>(null);

// Auth state
export const isLoggedIn = writable(false);

// Subscribe to store changes to update localStorage
if (browser) {
  isDarkMode.subscribe(value => {
    if (value) {
      document.documentElement.classList.add('dark');
      localStorage.setItem('theme', 'dark');
    } else {
      document.documentElement.classList.remove('dark');
      localStorage.setItem('theme', 'light');
    }
  });

  themePreference.subscribe(value => {
    localStorage.setItem('themePreference', value);
  });

  activeSessionId.subscribe(value => {
    if (value) {
      localStorage.setItem('activeSessionId', value);
    } else {
      localStorage.removeItem('activeSessionId');
    }
  });

  activeModel.subscribe(value => {
    if (value) {
      localStorage.setItem('activeModel', value);
    } else {
      localStorage.removeItem('activeModel');
    }
  });
} 