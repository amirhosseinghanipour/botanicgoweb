<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { auth } from '$lib/stores/auth';
  import { api } from '$lib/api/client';
  import { notifications } from '$lib/stores/notifications';
  import type { Message } from '$lib/types';

  let messages: Message[] = [];
  let isLoading = true;
  let error: string | null = null;
  let newMessage = '';
  let isSending = false;

  let sessionId: string;

  $: sessionId = $page.params.sessionId;

  async function loadSession() {
    try {
      const session = await api.getSession(sessionId);
      messages = session.messages;
    } catch (err) {
      console.error('Failed to load session:', err);
      error = 'Failed to load chat session';
      notifications.add({
        type: 'error',
        message: 'Failed to load chat session',
        duration: 5000
      });
    } finally {
      isLoading = false;
    }
  }

  async function sendMessage() {
    if (!newMessage.trim() || isSending) return;

    isSending = true;
    try {
      const response = await api.sendMessage(sessionId, newMessage);
      messages = [...messages, response];
      newMessage = '';
    } catch (err) {
      console.error('Failed to send message:', err);
      notifications.add({
        type: 'error',
        message: 'Failed to send message',
        duration: 5000
      });
    } finally {
      isSending = false;
    }
  }

  onMount(() => {
    if (!$auth.user) {
      goto('/login');
      return;
    }
    loadSession();
  });
</script>

<div class="flex flex-col h-[calc(100vh-4rem)]">
  {#if isLoading}
    <div class="flex-1 flex items-center justify-center">
      <div class="text-center">
        <svg class="w-8 h-8 animate-spin mx-auto mb-4" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M12 4.75V6.25" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
          <path d="M17.1475 6.8525L16.0625 7.9375" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
          <path d="M19.25 12H17.75" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
          <path d="M17.1475 17.1475L16.0625 16.0625" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
          <path d="M12 17.75V19.25" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
          <path d="M6.8525 17.1475L7.9375 16.0625" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
          <path d="M4.75 12H6.25" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
          <path d="M6.8525 6.8525L7.9375 7.9375" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
        <h2 class="text-lg font-medium text-gray-900 dark:text-white">Loading Chat</h2>
      </div>
    </div>
  {:else if error}
    <div class="flex-1 flex items-center justify-center">
      <div class="text-center text-red-600 dark:text-red-400">
        <p>{error}</p>
      </div>
    </div>
  {:else}
    <div class="flex-1 overflow-y-auto p-4 space-y-4">
      {#each messages as message}
        <div class="flex {message.user_id === 'assistant' ? 'justify-start' : 'justify-end'}">
          <div class="max-w-[80%] rounded-lg p-4 {message.user_id === 'assistant' ? 'bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-white' : 'bg-blue-500 text-white'}">
            <p class="whitespace-pre-wrap">{message.content}</p>
          </div>
        </div>
      {/each}
    </div>

    <div class="border-t border-gray-200 dark:border-gray-700 p-4">
      <div class="flex gap-4">
        <textarea
          bind:value={newMessage}
          on:keydown={(e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
              e.preventDefault();
              sendMessage();
            }
          }}
          placeholder="Type your message..."
          class="flex-1 px-4 py-2 text-gray-900 dark:text-white bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
          rows="1"
        ></textarea>
        <button
          on:click={sendMessage}
          disabled={!newMessage.trim() || isSending}
          class="px-4 py-2 text-sm font-medium text-white bg-blue-500 rounded-lg hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
        >
          {#if isSending}
            <div class="flex items-center gap-2">
              <svg class="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 4.75V6.25" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                <path d="M17.1475 6.8525L16.0625 7.9375" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                <path d="M19.25 12H17.75" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                <path d="M17.1475 17.1475L16.0625 16.0625" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                <path d="M12 17.75V19.25" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                <path d="M6.8525 17.1475L7.9375 16.0625" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                <path d="M4.75 12H6.25" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                <path d="M6.8525 6.8525L7.9375 7.9375" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
              <span>Sending...</span>
            </div>
          {:else}
            Send
          {/if}
        </button>
      </div>
    </div>
  {/if}
</div> 