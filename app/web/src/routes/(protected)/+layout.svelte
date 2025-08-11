<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { dev } from '$app/environment';
  import { auth } from "$lib/stores/auth";

  let isLoading = true;

  onMount(async () => {
    if (dev) {
      // In dev mode, bypass auth guard to allow UI access
      isLoading = false;
      return;
    }
    try {
      if (!$auth.user) {
        const isAuthenticated = await auth.checkAuth();
        if (!isAuthenticated) {
          goto('/');
          return;
        }
      }
    } catch (error) {
      console.error('Auth check failed:', error);
      goto('/');
    } finally {
      isLoading = false;
    }
  });
</script>

{#if isLoading}
  <div class="flex items-center justify-center min-h-screen">
    <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-black dark:border-white"></div>
  </div>
{:else if dev || $auth.user}
  <slot />
{/if}

