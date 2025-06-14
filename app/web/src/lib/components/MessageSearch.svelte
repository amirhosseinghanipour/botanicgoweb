<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { api } from '$lib/api/client';
  import { debounce } from '$lib/utils/debounce';

  export let sessionId: string;
  export let isOpen = false;

  const dispatch = createEventDispatcher<{
    select: { message: any };
    close: void;
  }>();

  let searchQuery = '';
  let searchResults: any[] = [];
  let isLoading = false;
  let error: string | null = null;

  const handleSearch = debounce(async (query: string) => {
    if (!query.trim()) {
      searchResults = [];
      return;
    }

    isLoading = true;
    error = null;

    try {
      searchResults = await api.searchMessages(sessionId, query);
    } catch (err) {
      console.error('Search failed:', err);
      error = 'Failed to search messages';
    } finally {
      isLoading = false;
    }
  }, 300);

  const handleSelect = (message: any) => {
    dispatch('select', { message });
    isOpen = false;
  };

  const handleClose = () => {
    dispatch('close');
    isOpen = false;
    searchQuery = '';
    searchResults = [];
  };

  $: if (searchQuery) {
    handleSearch(searchQuery);
  }
</script>

{#if isOpen}
  <div class="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center">
    <div class="bg-white dark:bg-neutral-900/50 rounded-lg shadow-xl w-full max-w-2xl max-h-[80vh] flex flex-col">
      <div class="p-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">Search Messages</h2>
        <button
          class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
          on:click={handleClose}
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="p-4">
        <input
          type="text"
          class="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-neutral-900/50 dark:text-white"
          placeholder="Search messages..."
          bind:value={searchQuery}
        />
      </div>

      <div class="flex-1 overflow-y-auto p-4">
        {#if isLoading}
          <div class="flex justify-center py-4">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
          </div>
        {:else if error}
          <div class="text-red-500 text-center py-4">{error}</div>
        {:else if searchResults.length === 0 && searchQuery}
          <div class="text-gray-500 dark:text-gray-400 text-center py-4">No messages found</div>
        {:else}
          <div class="space-y-4">
            {#each searchResults as result (result.id)}
              <button
                class="w-full text-left p-4 rounded-lg hover:bg-gray-100 dark:hover:bg-neutral-900/50 transition-colors"
                on:click={() => handleSelect(result)}
              >
                <div class="text-sm text-gray-500 dark:text-gray-400">
                  {new Date(result.createdAt).toLocaleString()}
                </div>
                <div class="mt-1 text-gray-900 dark:text-white">
                  {result.content}
                </div>
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if} 