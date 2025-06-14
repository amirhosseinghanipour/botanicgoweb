<script lang="ts">
  import { theme } from '$lib/stores/theme';
  import { icons } from '$lib/icons';
</script>

<div class="relative">
  <button
    class="flex items-center justify-center w-10 h-10 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800"
    aria-label="Toggle theme"
    on:click={() => theme.toggle()}
  >
    {#if $theme === 'dark'}
      {@html icons.sun}
    {:else if $theme === 'light'}
      {@html icons.moon}
    {:else}
      {@html icons.system}
    {/if}
  </button>
</div> 