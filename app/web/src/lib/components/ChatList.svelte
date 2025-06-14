<script lang="ts">
  import { sessions, activeSessionId } from "$lib/stores";
  import { api } from "$lib/api/client";
  import { notifications } from '$lib/stores/notifications';
  import { goto } from '$app/navigation';
  import { fade } from 'svelte/transition';

  let isCreating = false;
  let searchQuery = '';
  let showSearch = false;

  $: filteredSessions = searchQuery
    ? $sessions.filter(session => 
        session.title.toLowerCase().includes(searchQuery.toLowerCase())
      )
    : $sessions;

  async function createSession() {
    if (isCreating) return;
    isCreating = true;

    try {
      const session = await api.createSession({
        title: 'New Chat',
        model: 'DeepSeek: DeepSeek V3 (free)'
      });

      $sessions = [
        {
          id: session.id,
          user_id: session.user_id,
          title: session.title,
          model: session.model || 'default',
          created_at: session.created_at,
          updated_at: session.updated_at,
          messages: []
        },
        ...$sessions
      ];

      $activeSessionId = session.id;
    } catch (error) {
      console.error('Failed to create session:', error);
      notifications.add({
        type: 'error',
        message: 'Failed to create new chat',
        duration: 5000
      });
    } finally {
      isCreating = false;
    }
  }

  function selectSession(session: typeof $sessions[0]) {
    $activeSessionId = session.id;
  }

  function formatDate(date: string) {
    const d = new Date(date);
    const now = new Date();
    const diff = now.getTime() - d.getTime();
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));

    if (days === 0) {
      return 'Today';
    } else if (days === 1) {
      return 'Yesterday';
    } else if (days < 7) {
      return `${days} days ago`;
    } else {
      return d.toLocaleDateString();
    }
  }
</script>

<div class="flex flex-col h-full">
  <div class="flex flex-col gap-2 p-4 border-b border-neutral-200 dark:border-neutral-800">
    <div class="flex items-center justify-between">
      <h2 class="text-lg font-semibold text-black dark:text-white">Chats</h2>
      <div class="flex items-center gap-2">
        <button
          class="p-2 text-neutral-600 dark:text-neutral-400 hover:text-black dark:hover:text-white"
          on:click={() => showSearch = !showSearch}
          title="Search chats"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="w-5 h-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            />
          </svg>
        </button>
        <button
          class="p-2 text-neutral-600 dark:text-neutral-400 hover:text-black dark:hover:text-white"
          on:click={createSession}
          disabled={isCreating}
          title="New chat"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="w-5 h-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 4v16m8-8H4"
            />
          </svg>
        </button>
      </div>
    </div>

    {#if showSearch}
      <div class="relative" in:fade={{ duration: 200 }}>
        <input
          type="text"
          bind:value={searchQuery}
          placeholder="Search chats..."
          class="w-full px-4 py-2 text-sm bg-neutral-100 dark:bg-neutral-800 border border-neutral-200 dark:border-neutral-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        {#if searchQuery}
          <button
            class="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-neutral-400 hover:text-neutral-600 dark:hover:text-neutral-200"
            on:click={() => searchQuery = ''}
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="w-4 h-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        {/if}
      </div>
    {/if}
  </div>

  <div class="flex-1 overflow-y-auto p-2 space-y-1">
    {#each filteredSessions as session (session.id)}
      <button
        class="w-full p-3 text-left rounded-lg transition-colors
          {$activeSessionId === session.id
            ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400'
            : 'hover:bg-neutral-100 dark:hover:bg-neutral-800 text-neutral-900 dark:text-neutral-100'}"
        on:click={() => selectSession(session)}
      >
        <div class="flex items-center justify-between">
          <span class="font-medium truncate">{session.title}</span>
          <span class="text-sm opacity-60">
            {formatDate(session.updatedAt)}
          </span>
        </div>
      </button>
    {/each}
  </div>
</div> 