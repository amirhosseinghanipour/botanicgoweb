<script lang="ts">
import { isOffline } from '$lib/stores/offline';
import { derived } from 'svelte/store';

const show = derived(isOffline, $isOffline => $isOffline);
</script>

{#if $show}
  <div class="fixed top-0 left-0 w-full z-50 bg-red-600 text-white text-center py-2 shadow-lg">
    <span>You are offline. Some features may not work.</span>
  </div>
{/if} 