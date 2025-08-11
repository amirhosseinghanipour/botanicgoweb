<script lang="ts">
  import "../app.css";

  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import { isLoggedIn } from "$lib/stores";
  import { icons } from "$lib/icons.js";
  import { auth } from '$lib/stores/auth';
  import { goto } from "$app/navigation";
  import { theme } from '$lib/stores/theme';
  import { browser } from '$app/environment';
  import OfflineBanner from '$lib/components/OfflineBanner.svelte';
  import Header from '$lib/components/layout/Header.svelte';

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
    // Theme store initializes itself; ensure it reads current value
    const storedTheme = localStorage.getItem('theme') as 'light' | 'dark' | 'system' | null;
    if (storedTheme) theme.set(storedTheme);
  });

  $: isLoggedIn.set(!!$auth.user);

  // Theme changes handled by $lib/stores/theme

  const handleLogout = () => {
    auth.logout();
    goto('/');
  };
</script>

<svelte:head>
  <script>
    // Prevent theme flash
    (function() {
        const t = localStorage.getItem('theme') || 'system';
        const isDark = t === 'dark' || (t === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);
        if (isDark) {
          document.documentElement.classList.add('dark');
          document.documentElement.style.colorScheme = 'dark';
        } else {
          document.documentElement.classList.remove('dark');
          document.documentElement.style.colorScheme = t === 'system' ? 'light dark' : 'light';
        }
    })();
  </script>
</svelte:head>

<OfflineBanner />

<div
  class="bg-white dark:bg-black text-black dark:text-white min-h-screen font-sans flex flex-col selection-style"
>
  {#if !$page.url.pathname.startsWith('/chat')}
    <Header />
  {/if}

  <main class="px-4 flex-grow">
    <slot />
  </main>

  {#if !$page.url.pathname.startsWith('/chat')}
    <footer
      class="fixed bottom-0 left-0 w-full glass z-50 border-t border-white/40 dark:border-white/10"
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
  {/if}
</div>

<style>
  :global(body) {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
      "Helvetica Neue", Arial, sans-serif;
  }
</style>