<script lang="ts">
  import { theme } from '$lib/stores/theme';
  import { icons } from '$lib/icons';

  let isOpen = false;

  function handleThemeSelect(newTheme: 'light' | 'dark' | 'system') {
    theme.set(newTheme);
    isOpen = false;
  }

  function handleClickOutside(event: MouseEvent) {
    const target = event.target as HTMLElement;
    if (!target.closest('.theme-selector')) {
      isOpen = false;
    }
  }
</script>

<svelte:window on:click={handleClickOutside} />

<div class="theme-selector relative">
  <button
    class="flex items-center justify-center w-10 h-10 rounded-lg hover:bg-black/5 dark:hover:bg-white/5"
    aria-label="Select theme"
    on:click={() => (isOpen = !isOpen)}
  >
    {#if $theme === 'dark'}
      {@html icons.sun}
    {:else if $theme === 'light'}
      {@html icons.moon}
    {:else}
      {@html icons.system}
    {/if}
  </button>

  {#if isOpen}
    <div
      class="absolute right-0 mt-2 w-48 rounded-xl glass z-50 focus:outline-none"
      role="menu"
      aria-orientation="vertical"
      aria-labelledby="theme-menu"
    >
      <div class="py-1" role="none">
        <button
          class="flex items-center w-full px-4 py-2 text-sm text-gray-800 dark:text-gray-100 hover:bg-black/5 dark:hover:bg-white/10 rounded-lg"
          role="menuitem"
          on:click={() => handleThemeSelect('light')}
        >
          {@html icons.sun}
          <span class="ml-3">Light</span>
        </button>
        <button
          class="flex items-center w-full px-4 py-2 text-sm text-gray-800 dark:text-gray-100 hover:bg-black/5 dark:hover:bg-white/10 rounded-lg"
          role="menuitem"
          on:click={() => handleThemeSelect('dark')}
        >
          {@html icons.moon}
          <span class="ml-3">Dark</span>
        </button>
        <button
          class="flex items-center w-full px-4 py-2 text-sm text-gray-800 dark:text-gray-100 hover:bg-black/5 dark:hover:bg-white/10 rounded-lg"
          role="menuitem"
          on:click={() => handleThemeSelect('system')}
        >
          {@html icons.system}
          <span class="ml-3">System</span>
        </button>
      </div>
    </div>
  {/if}
</div> 