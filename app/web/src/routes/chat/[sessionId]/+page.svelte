<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import {
    store as websocketStore,
    messages,
    isConnected,
    isLoading,
    error,
  } from "$lib/stores/websocket";
  import Message from "$lib/components/Message.svelte";
  import ChatInput from "$lib/components/ChatInput.svelte";

  let a_isComponentMounted = false;

  onMount(() => {
    const sessionId = $page.params.sessionId;
    if (sessionId) {
      // Connect to the WebSocket and load initial messages
      websocketStore.connect(sessionId);
      websocketStore.loadMessages(sessionId);
    }
    a_isComponentMounted = true;

    return () => {
      // Disconnect when the component is destroyed
      websocketStore.disconnect();
    };
  });
</script>

<div class="flex flex-col h-screen bg-gray-100 dark:bg-gray-900">
  <header class="bg-white dark:bg-gray-800 shadow-md p-4">
    <h1 class="text-xl font-bold text-gray-800 dark:text-white">
      Chat Session: {$page.params.sessionId}
    </h1>
    <div class="text-sm text-gray-500 dark:text-gray-400">
      {#if $isConnected}
        <span class="text-green-500">Connected</span>
      {:else if $isLoading}
        <span>Connecting...</span>
      {:else}
        <span class="text-red-500">Disconnected</span>
      {/if}
    </div>
  </header>

  <main class="flex-1 overflow-y-auto p-4">
    <div class="space-y-4">
      {#if a_isComponentMounted}
        {#each $messages as message (message.id)}
          <Message {message} />
        {/each}
      {/if}

      {#if $isLoading && $messages.length === 0}
        <div class="text-center text-gray-500 dark:text-gray-400">
          Loading messages...
        </div>
      {/if}

      {#if $error}
        <div class="text-center text-red-500">
          Error: {$error}
        </div>
      {/if}
    </div>
  </main>

  <footer
    class="p-4 bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700"
  >
    <ChatInput />
  </footer>
</div>

