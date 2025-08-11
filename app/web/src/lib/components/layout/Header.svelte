<script lang="ts">
  import { goto } from "$app/navigation";
  import { auth } from "$lib/stores/auth";
  import ThemeSelector from "$lib/components/ThemeSelector.svelte";

  const handleLogout = () => {
    auth.logout();
    goto('/');
  };
</script>

<header class="sticky top-0 z-30 glass border-b border-white/40 dark:border-white/10 backdrop-blur-2xl saturate-150">
  <div class="h-14 md:h-16 px-4 md:px-8 flex items-center justify-between">
    
    <div class="flex items-center gap-3">
      <a href="/" class="font-semibold tracking-tight text-sm md:text-base text-gray-900 dark:text-gray-100">BOTANIC</a>

      
      <nav class="hidden md:flex items-center gap-1 text-sm font-medium text-gray-600 dark:text-gray-300">
        {#if $auth.user}
          <a href="/settings/profile" class="px-3 py-1.5 rounded-full hover:bg-black/5 dark:hover:bg-white/5 hover:text-black dark:hover:text-white">Settings</a>
        {/if}
      </nav>
    </div>

    
    <div class="flex items-center gap-2">
      
      {#if $auth.user}
        <a href="/settings/profile" class="hidden md:inline-flex px-3 py-1.5 rounded-full text-sm font-medium text-gray-600 dark:text-gray-300 hover:bg-black/5 dark:hover:bg-white/5 hover:text-black dark:hover:text-white">Settings</a>
      {/if}

      
      <ThemeSelector />

      {#if $auth.user}
        <button
          on:click={handleLogout}
          class="hidden sm:inline-flex items-center gap-2 px-4 py-1.5 text-sm font-bold bg-black dark:bg-white text-white dark:text-black rounded-full shadow-sm hover:opacity-90 transition-opacity"
        >
          Logout
        </button>
      {:else}
        <a
          href="/login"
          class="hidden sm:inline-flex items-center gap-2 px-3 py-1.5 text-sm font-medium rounded-full border border-black/10 dark:border-white/10 text-gray-800 dark:text-gray-100 hover:bg-black/5 dark:hover:bg-white/10 transition-colors"
        >
         Sign in
        </a>
        <a
          href="/register"
          class="hidden sm:inline-flex items-center gap-2 px-4 py-1.5 text-sm font-bold rounded-full text-white bg-black dark:bg-white dark:text-black shadow-sm hover:opacity-90 transition-opacity"
        >
          Sign up
        </a>
      {/if}
    </div>
  </div>
</header>