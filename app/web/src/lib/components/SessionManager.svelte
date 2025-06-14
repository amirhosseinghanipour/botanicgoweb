<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import { goto } from '$app/navigation';
  import type { SessionInfo } from '$lib/api/client';

  let sessions: SessionInfo[] = [];
  let loading = true;
  let error: string | null = null;

  onMount(async () => {
    try {
      sessions = await api.getUserSessions();
    } catch (err) {
      error = 'Failed to load sessions';
      console.error('Failed to load sessions:', err);
    } finally {
      loading = false;
    }
  });

  const handleDeleteSession = async (sessionId: string) => {
    try {
      await api.deleteUserSession(sessionId);
      sessions = sessions.filter(s => s.id !== sessionId);
    } catch (err) {
      error = 'Failed to delete session';
      console.error('Failed to delete session:', err);
    }
  };

  const formatDate = (date: string) => {
    return new Date(date).toLocaleString();
  };

  const getDeviceInfo = (userAgent: string) => {
    // Basic device detection
    if (userAgent.includes('Mobile')) {
      return 'Mobile Device';
    } else if (userAgent.includes('Tablet')) {
      return 'Tablet';
    } else {
      return 'Desktop';
    }
  };
</script>

<div class="space-y-4">
  <h2 class="text-lg font-medium text-gray-900 dark:text-gray-100">Active Sessions</h2>

  {#if loading}
    <div class="flex justify-center">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 dark:border-white"></div>
    </div>
  {:else if error}
    <div class="text-red-500 dark:text-red-400">{error}</div>
  {:else if sessions.length === 0}
    <p class="text-gray-500 dark:text-gray-400">No active sessions</p>
  {:else}
    <div class="space-y-4">
      {#each sessions as session (session.id)}
        <div class="flex items-center justify-between p-4 bg-white dark:bg-neutral-900/50 rounded-lg shadow">
          <div class="space-y-1">
            <p class="text-sm font-medium text-gray-900 dark:text-gray-100">
              {getDeviceInfo(session.device)}
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              Location: {session.location}
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              Created: {formatDate(session.createdAt)}
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              Expires: {formatDate(session.expiresAt)}
            </p>
          </div>
          <button
            class="p-2 text-red-500 hover:text-red-600 dark:text-red-400 dark:hover:text-red-300"
            on:click={() => handleDeleteSession(session.id)}
          >
            <svg
              class="w-5 h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
              />
            </svg>
          </button>
        </div>
      {/each}
    </div>
  {/if}
</div> 