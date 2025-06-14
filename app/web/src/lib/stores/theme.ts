import { browser } from '$app/environment';
import { writable } from 'svelte/store';

export type Theme = 'light' | 'dark' | 'system';

function createThemeStore() {
  const storedTheme = browser ? localStorage.getItem('theme') as Theme : 'system';
  const { subscribe, set, update } = writable<Theme>(storedTheme);

  return {
    subscribe,
    set: (theme: Theme) => {
      if (browser) {
        localStorage.setItem('theme', theme);
        updateTheme(theme);
      }
      set(theme);
    },
    toggle: () => {
      update((current) => {
        const newTheme = current === 'dark' ? 'light' : 'dark';
        if (browser) {
          localStorage.setItem('theme', newTheme);
          updateTheme(newTheme);
        }
        return newTheme;
      });
    },
  };
}

function updateTheme(theme: Theme) {
  const root = window.document.documentElement;
  const isDark = theme === 'dark' || (theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);
  
  root.classList.remove('light', 'dark');
  root.classList.add(isDark ? 'dark' : 'light');
}

export const theme = createThemeStore();

// Initialize theme on load
if (browser) {
  const storedTheme = localStorage.getItem('theme') as Theme;
  updateTheme(storedTheme || 'system');

  // Listen for system theme changes
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
    const theme = localStorage.getItem('theme') as Theme;
    if (theme === 'system') {
      updateTheme('system');
    }
  });
} 