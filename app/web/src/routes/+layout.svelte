<script lang="ts">
  import "../app.css";

  import { onMount } from "svelte";
  import {
    isDarkMode,
    themePreference,
    isLoggedIn,
    type ThemePreference
  } from "$lib/stores";
  import { icons } from "$lib/icons.js";
  import { auth } from '$lib/stores/auth';
  import { goto } from "$app/navigation";
  import { theme } from '$lib/stores/theme';
  import { browser } from '$app/environment';
  import OfflineBanner from '$lib/components/OfflineBanner.svelte';

  let isLoading = true;

  async function initializeAuth() {
    try {
      await auth.initializeAuth();
    } catch (err) {
      console.error("Failed to initialize auth:", err);
    } finally {
      isLoading = false;
    }
  }

  onMount(() => {
    initializeAuth();
    const savedTheme = localStorage.getItem("themePreference") as ThemePreference;
    if (savedTheme) {
      themePreference.set(savedTheme);
    }

    themePreference.subscribe((value) => {
      localStorage.setItem("themePreference", value);
    });

    isDarkMode.subscribe((value) => {
      document.documentElement.classList.toggle("dark", value);
    });

    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const handleSystemThemeChange = (e: MediaQueryListEvent) => {
      if ($themePreference === "system") {
        isDarkMode.set(e.matches);
      }
    };
    mediaQuery.addEventListener("change", handleSystemThemeChange);

    const storedTheme = localStorage.getItem('theme') as 'light' | 'dark' | 'system';
    if (storedTheme) {
      theme.set(storedTheme);
    }

    return () => mediaQuery.removeEventListener("change", handleSystemThemeChange);
  });

  $: isLoggedIn.set(!!$auth.user);

  const handleThemeChange = (value: ThemePreference) => {
    themePreference.set(value);
    if (value === "system") {
      isDarkMode.set(window.matchMedia("(prefers-color-scheme: dark)").matches);
    } else {
      isDarkMode.set(value === "dark");
    }
  };

  const handleLogout = () => {
    auth.logout();
    goto('/');
  };
</script>

<svelte:head>
  <script>
    // Prevent theme flash
    (function() {
      const theme = localStorage.getItem('themePreference') || 'system';
      const isDark = theme === 'dark' || (theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);
      if (isDark) {
        document.documentElement.classList.add('dark');
      }
    })();
  </script>
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
  <link
    href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap"
    rel="stylesheet"
  />
</svelte:head>

<OfflineBanner />

<div
  class="bg-white dark:bg-black text-black dark:text-white min-h-screen font-sans flex flex-col"
>
  <header
    class="h-16 flex items-center justify-between px-4 md:px-8 sticky top-0 bg-white/70 dark:bg-black/70 backdrop-blur-md z-20"
  >
    <div class="flex items-center gap-4">
      <a
        href="/"
        class="font-extrabold text-lg tracking-wide"
      >
        BOTANIC
      </a>
    </div>
    <div class="flex items-center gap-4">
      <nav
        class="hidden md:flex items-center gap-4 text-sm font-medium text-gray-500 dark:text-gray-400"
      >
        <a
          href="/features"
          class="px-3 py-1.5 hover:text-black dark:hover:text-white transition-colors"
        >
          Features
        </a>
        <a
          href="/pricing"
          class="px-3 py-1.5 hover:text-black dark:hover:text-white transition-colors"
        >
          Pricing
        </a>
        {#if $auth.user}
          <a
            href="/settings/profile"
            class="px-3 py-1.5 hover:text-black dark:hover:text-white transition-colors"
          >
            Settings
          </a>
        {/if}
      </nav>
      <div class="px-4 py-2">
        <select
          id="theme-select"
          bind:value={$themePreference}
          on:change={(e: Event) => handleThemeChange((e.target as HTMLSelectElement).value as ThemePreference)}
          class="pl-1 py-1 text-sm bg-gray-200 dark:bg-gray-800 rounded-md text-gray-700 dark:text-gray-300 border-gray-300 dark:border-gray-700"
        >
          <option value="system">System</option>
          <option value="light">Light</option>
          <option value="dark">Dark</option>
        </select>
      </div>
      {#if $auth.user}
        <button
          on:click={handleLogout}
          class="flex items-center gap-2 px-4 py-1.5 text-sm font-bold bg-black dark:bg-white text-white dark:text-black rounded-lg shadow-sm hover:opacity-90 transition-opacity"
        >
          Logout
        </button>
      {:else}
        <a
          href="/login"
          class="flex items-center gap-2 px-4 py-1.5 text-sm font-bold bg-black dark:bg-white text-white dark:text-black rounded-lg shadow-sm hover:opacity-90 transition-opacity"
        >
          Login
        </a>
      {/if}
    </div>
  </header>

  <main class="px-4 flex-grow">
    <slot />
  </main>

  <footer
    class="fixed bottom-0 left-0 w-full bg-white/70 dark:bg-black/70 backdrop-blur-md z-50 border-t border-gray-200 dark:border-gray-800"
  >
    <div
      class="flex items-center justify-between h-14 px-4 md:px-8 text-xs text-gray-500 dark:text-gray-400"
    >
      <div class="flex items-center gap-4">
        <span>Follow us on</span>
        <a
          href="https://x.com/botanic"
          class="hover:text-black dark:hover:text-white transition-colors"
          aria-label="Follow on X"
        >
          {@html icons.x}
        </a>
      </div>
      <div>
        By using Botanic you agree to the
        <a
          href="/terms"
          class="font-medium hover:text-black dark:hover:text-white transition-colors"
        >
          Terms
        </a>
        and
        <a
          href="/privacy"
          class="font-medium hover:text-black dark:hover:text-white transition-colors"
        >
          Privacy
        </a>.
      </div>
    </div>
  </footer>
</div>

<style>
  :global(body) {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
      "Helvetica Neue", Arial, sans-serif;
  }
</style>