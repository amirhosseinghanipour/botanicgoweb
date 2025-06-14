<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { auth } from "$lib/stores/auth";

  let isLoading = true;

  onMount(async () => {
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
{:else if $auth.user}
  <slot />
{/if}

