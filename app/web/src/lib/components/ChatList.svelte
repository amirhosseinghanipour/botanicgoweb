<script lang="ts">
  import { sessions, activeSessionId } from "$lib/stores";
  import { api } from "$lib/api/client";
  import { notifications } from '$lib/stores/notifications';
  import { goto } from '$app/navigation';
  import { fade } from 'svelte/transition';
  import { llmStore } from '$lib/stores/llm';
  import { auth } from '$lib/stores/auth';
  import { theme } from '$lib/stores/theme';

  let showProfileDialog = false;
  let showSettingsDialog = false;
  let byokKey = '';

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
        model: $llmStore.selectedModel?.id || 'gpt-oss-20b'
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
  
  <div class="p-4 border-b border-neutral-200 dark:border-neutral-800">
    <div class="flex gap-2">
      <div class="relative flex-1">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-neutral-400" viewBox="0 0 24 24" fill="none" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/></svg>
        <input
          type="text"
          bind:value={searchQuery}
          placeholder="Search chats..."
          class="w-full pl-9 pr-3 py-2 text-xs rounded-lg bg-neutral-100 dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 focus:outline-none focus:ring-1 focus:ring-black dark:focus:ring-white"
        />
      </div>
      <button
        class="px-3 py-2 text-xs font-medium rounded-lg bg-black dark:bg-white text-white dark:text-black hover:bg-gray-800 dark:hover:bg-gray-200 transition-colors"
        on:click={createSession}
        disabled={isCreating}
        aria-label="New chat"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
      </button>
    </div>
  </div>

  <div class="flex-1 overflow-y-auto p-2 space-y-1.5">
    {#each filteredSessions as session (session.id)}
      <button
        class="w-full text-left rounded-lg transition-colors hover:bg-black/5 dark:hover:bg-white/5 px-2.5 py-2
          {$activeSessionId === session.id ? 'bg-black/5 dark:bg-white/5' : ''}"
        on:click={() => selectSession(session)}
      >
        <div class="flex items-center gap-2">
          <div class="w-6 h-6 rounded bg-neutral-200 dark:bg-neutral-800"></div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between gap-2">
              <span class="text-sm font-medium truncate">{session.title}</span>
              <span class="text-[11px] text-neutral-500 shrink-0">{formatDate(session.updatedAt)}</span>
            </div>
            <div class="text-[11px] text-neutral-500 truncate">{session.model}</div>
          </div>
        </div>
      </button>
    {/each}
  </div>

  
  <div class="p-3 border-t border-neutral-200 dark:border-neutral-800">
    <div class="flex items-center gap-2">
      <div class="w-8 h-8 rounded-full bg-neutral-200 dark:bg-neutral-800 flex items-center justify-center text-xs font-medium">
        {#if $auth.user?.email}
          {($auth.user.email[0] || 'U').toUpperCase()}
        {:else}
          U
        {/if}
      </div>
      <div class="flex-1 min-w-0">
        <div class="text-sm font-medium truncate">{$auth.user?.name || 'User'}</div>
        <div class="text-xs text-neutral-500 truncate">{$auth.user?.email || 'Not signed in'}</div>
      </div>
      <button class="px-2 py-1 text-xs rounded-lg hover:bg-black/5 dark:hover:bg-white/10" on:click={() => showProfileDialog = true}>Profile</button>
      <button class="px-2 py-1 text-xs rounded-lg hover:bg-black/5 dark:hover:bg-white/10" on:click={() => showSettingsDialog = true}>Settings</button>
    </div>
  </div>

  {#if showProfileDialog}
    <div class="fixed inset-0 z-50 bg-black/40 backdrop-blur-md flex items-center justify-center" role="button" aria-label="Close" tabindex="0" on:click={() => showProfileDialog = false} on:keydown={(e) => (e.key === 'Escape' || e.key === 'Enter' || e.key === ' ') && (showProfileDialog = false)}>
      <div class="w-full max-w-md mx-4 glass rounded-2xl border border-white/10 p-6" role="dialog" aria-modal="true" tabindex="-1" on:click|stopPropagation>
        <div class="flex items-center gap-3 mb-4">
          <button class="p-2 rounded-lg hover:bg-neutral-100 dark:hover:bg-neutral-800 transition-colors" aria-label="Back" on:click={() => showProfileDialog = false}>
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
          </button>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Your Profile</h3>
        </div>
        <div class="space-y-4">
          <div>
            <label for="profile_name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Name</label>
            <input id="profile_name" class="w-full px-3 py-2 text-sm rounded-lg bg-neutral-100 dark:bg-neutral-800 border border-neutral-200 dark:border-neutral-700 text-gray-900 dark:text-gray-100" value={$auth.user?.name || ''} disabled />
          </div>
          <div>
            <label for="profile_email" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Email</label>
            <input id="profile_email" class="w-full px-3 py-2 text-sm rounded-lg bg-neutral-100 dark:bg-neutral-800 border border-neutral-200 dark:border-neutral-700 text-gray-900 dark:text-gray-100" value={$auth.user?.email || ''} disabled />
          </div>
          <div class="pt-2 text-sm text-gray-500 dark:text-gray-400 bg-neutral-50 dark:bg-neutral-800 p-3 rounded-lg">
            Managed by Botanic. Profile editing is coming soon.
          </div>
        </div>
        <div class="mt-6 flex justify-end">
          <button class="px-4 py-2 text-sm font-medium rounded-lg bg-black dark:bg-white text-white dark:text-black hover:bg-gray-800 dark:hover:bg-gray-200 transition-colors" on:click={() => showProfileDialog = false}>Close</button>
        </div>
      </div>
    </div>
  {/if}

  {#if showSettingsDialog}
    <div class="fixed inset-0 z-50 bg-black/40 backdrop-blur-md flex items-center justify-center" role="button" aria-label="Close" tabindex="0" on:click={() => showSettingsDialog = false} on:keydown={(e) => (e.key === 'Escape' || e.key === 'Enter' || e.key === ' ') && (showSettingsDialog = false)}>
      <div class="w-full max-w-lg mx-4 glass rounded-2xl border border-white/10 p-6" role="dialog" aria-modal="true" tabindex="-1" on:click|stopPropagation>
        <div class="flex items-center gap-3 mb-4">
          <button class="p-2 rounded-lg hover:bg-neutral-100 dark:hover:bg-neutral-800 transition-colors" aria-label="Back" on:click={() => showSettingsDialog = false}>
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
          </button>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Settings</h3>
        </div>
        <div class="space-y-6">
          <div>
            <h4 class="text-base font-medium text-gray-900 dark:text-gray-100 mb-3">General</h4>
            <div class="space-y-3">
              <div class="flex items-center justify-between">
                <span class="text-sm text-gray-700 dark:text-gray-300">Theme</span>
                <div class="flex items-center gap-1">
                  <button class="px-3 py-1.5 text-xs rounded-lg border border-neutral-200 dark:border-neutral-700 hover:bg-neutral-50 dark:hover:bg-neutral-800 transition-colors {$theme === 'light' ? 'bg-black text-white dark:bg-white dark:text-black' : 'bg-transparent'}" on:click={() => theme.set('light')}>Light</button>
                  <button class="px-3 py-1.5 text-xs rounded-lg border border-neutral-200 dark:border-neutral-700 hover:bg-neutral-50 dark:hover:bg-neutral-800 transition-colors {$theme === 'dark' ? 'bg-black text-white dark:bg-white dark:text-black' : 'bg-transparent'}" on:click={() => theme.set('dark')}>Dark</button>
                  <button class="px-3 py-1.5 text-xs rounded-lg border border-neutral-200 dark:border-neutral-700 hover:bg-neutral-50 dark:hover:bg-neutral-800 transition-colors {$theme === 'system' ? 'bg-black text-white dark:bg-white dark:text-black' : 'bg-transparent'}" on:click={() => theme.set('system')}>System</button>
                </div>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm text-gray-700 dark:text-gray-300">Notifications</span>
                <span class="text-xs text-gray-500 dark:text-gray-400 bg-neutral-50 dark:bg-neutral-800 px-2 py-1 rounded">Coming soon</span>
              </div>
            </div>
          </div>
          <div>
            <h4 class="text-base font-medium text-gray-900 dark:text-gray-100 mb-3">Integrations</h4>
            <div class="space-y-3">
              <div>
                <label for="byok" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Bring Your Own Key (BYOK)</label>
                <input id="byok" class="w-full px-3 py-2 text-sm rounded-lg bg-neutral-100 dark:bg-neutral-800 border border-neutral-200 dark:border-neutral-700 text-gray-900 dark:text-gray-100" bind:value={byokKey} placeholder="sk-..." />
                <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">Stored locally. Used when calling third-party providers.</div>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm text-gray-700 dark:text-gray-300">Data export</span>
                <span class="text-xs text-gray-500 dark:text-gray-400 bg-neutral-50 dark:bg-neutral-800 px-2 py-1 rounded">Coming soon</span>
              </div>
            </div>
          </div>
          <div>
            <h4 class="text-base font-medium text-gray-900 dark:text-gray-100 mb-3">Danger zone</h4>
            <div class="space-y-3">
              <button class="w-full px-4 py-2 text-sm font-medium rounded-lg border border-red-600/30 text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors" on:click={() => goto('/settings/profile')}>Delete account</button>
            </div>
          </div>
        </div>
        <div class="mt-6 flex justify-end">
          <button class="px-4 py-2 text-sm font-medium rounded-lg bg-black dark:bg-white text-white dark:text-black hover:bg-gray-800 dark:hover:bg-gray-200 transition-colors" on:click={() => showSettingsDialog = false}>Close</button>
        </div>
      </div>
    </div>
  {/if}
</div> 