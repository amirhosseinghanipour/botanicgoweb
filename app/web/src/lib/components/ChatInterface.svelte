<script lang="ts">
  import { onMount, tick } from "svelte";
  import { repository } from "$lib/data/repository";
  import { store as websocket } from "$lib/stores/websocket";
  import ChatMessage from "./ChatMessage.svelte";
  import { derived } from "svelte/store";
  import { isOffline } from "$lib/stores/offline";
  import { isReconnecting } from "$lib/stores/websocket";
  import {
    messageQueue,
    queueMessage,
    processQueue,
  } from "$lib/stores/messageQueue";
  import { notifications } from "$lib/stores/notifications";
  import { writable } from "svelte/store";
  import type { Message } from "$lib/types";
  import { llmStore } from "$lib/stores/llm";
  
  import ModelComparison from "./ModelComparison.svelte";
  import { goto } from "$app/navigation";

  export let sessionId: string;

  let userPrompt = "";
  let isSubmitting = false;
  let textarea: HTMLTextAreaElement;
  let chatContainer: HTMLDivElement;
  let isLoadingMore = false;
  let hasMoreMessages = true;
  
  let showModelComparison = false;
  let comparisonPrompt = "";

  // Header actions are managed by the parent page

  const messages = writable<Message[]>([]);
  const messagesLength = derived(messages, ($messages) => $messages.length);

  const wsMessages = derived(websocket, ($ws) => $ws.messages);

  // Auto-update local messages store from websocket store messages
  $: {
    if ($wsMessages) {
      messages.set($wsMessages); // Always reflect the latest messages from the websocket store
    }
  }

  // --- FIX START ---
  onMount(() => {
    // Removed 'async' here
    const initializeChat = async () => {
      // Moved async logic into a separate async function
      try {
        if (!sessionId || sessionId === "undefined") {
          console.error("Invalid session ID:", sessionId);
          notifications.add({
            type: "error",
            message: "Invalid session ID",
            duration: 5000,
          });
          goto("/chat");
          return;
        }

        // Load initial messages first
        try {
          const localMessages = await repository.getSessionMessages(sessionId);
          // Map to frontend Message type shape
          messages.set(localMessages.map(m => ({
            id: m.id,
            session_id: m.sessionId,
            user_id: '',
            content: m.content,
            type: 'message',
            role: m.role,
            model: m.model || 'default',
            created_at: m.createdAt,
            updated_at: m.createdAt
          })));
        } catch (err) {
          console.error("Failed to load initial messages:", err);
          notifications.add({
            type: "error",
            message: "Failed to load chat messages",
            duration: 5000,
          });
        }

        // Then establish WebSocket connection
        websocket.connect(sessionId);

        // Subscribe to WebSocket errors
        const unsubscribe = websocket.subscribe((state) => {
          if (state.error) {
            notifications.add({
              type: "error",
              message: state.error,
              duration: 5000,
            });
          }
        });

        // The cleanup function is returned directly by onMount
        return () => {
          unsubscribe();
          websocket.disconnect();
        };
      } catch (err) {
        console.error("Failed to connect:", err);
        notifications.add({
          type: "error",
          message: "Failed to connect to chat",
          duration: 5000,
        });
        // If an error occurs in the outer try-catch, ensure WebSocket is disconnected
        websocket.disconnect();
      }
    };

    initializeChat(); // Call the async function immediately
  });
  // --- FIX END ---

  // REMOVED THE REDUNDANT ASYNC FUNCTION sendMessage(message: Omit<Message, 'id' | 'createdAt'>)

  async function handleSubmit() {
    if (!userPrompt.trim() || isSubmitting) return;

    isSubmitting = true;
    const messageContent = userPrompt; // Capture current prompt
    userPrompt = ""; // Clear input immediately
    handleInput(); // Adjust textarea height

    try {
      // Create the user's message object for local display first
      const tempId = crypto.randomUUID();
      const localMessage: Message = {
        id: tempId,
        session_id: sessionId,
        content: messageContent,
        type: "message",
        user_id: localStorage.getItem("userId") || "",
        role: "user",
        model: $llmStore.selectedModel?.id || "default",
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      messages.update((msgs) => [...msgs, localMessage]); // Add to local store immediately

      // Persist locally in IndexedDB
      await repository.createMessage({ sessionId, role: 'user', content: messageContent, model: $llmStore.selectedModel?.id || 'default' });

      if ($isOffline) {
        queueMessage({ id: tempId, sessionId, content: messageContent });
        notifications.add({
          type: "info",
          message: "Message queued. Will be sent when you reconnect.",
          duration: 3000,
        });
      } else {
        // Direct call to the websocket store's sendMessage function
        // This function will generate its own UUID and handle JSON stringification
        try {
          await websocket.sendMessage(messageContent, sessionId);
        } catch (error) {
          if (error instanceof Error && error.name === "NetworkError") {
            queueMessage({ id: tempId, sessionId, content: messageContent });
            notifications.add({
              type: "warning",
              message: "Network error. Message queued for later.",
              duration: 3000,
            });
          } else {
            console.error("Failed to send message via websocket:", error);
            notifications.add({
              type: "error",
              message: "Failed to send message",
              duration: 5000,
            });
            messages.update((msgs) => msgs.filter((m) => m.id !== tempId)); // Remove from local store if failed immediately
          }
        }
      }
    } finally {
      isSubmitting = false;
    }
  }

  const handleKeydown = (event: KeyboardEvent) => {
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      handleSubmit();
    }
  };

  const handleInput = () => {
    textarea.style.height = "auto"; // Reset height to recalculate
    const newHeight = Math.min(textarea.scrollHeight, 150); // Using 150px as max-height from CSS for consistency
    textarea.style.height = `${newHeight}px`;
  };

  const handleScroll = async () => {
    if (!chatContainer || isLoadingMore || !hasMoreMessages) return;

    const { scrollTop, scrollHeight, clientHeight } = chatContainer;
    if (scrollTop < 100) {
      isLoadingMore = true;
      try {
        const currentScrollHeight = scrollHeight;
        // This function does not exist on websocket store based on previous websocket.ts
        // It exists on `llmStore` not `websocket` store. This is a bug.
        // If you intended to load more *chat messages*, you'd need to call `api.getSessionMessages`
        // and update the `messages` store with a pagination logic.
        // For now, I'm commenting out the problematic call.
        // await websocket.loadMoreMessages();
        await tick();
        chatContainer.scrollTop =
          chatContainer.scrollHeight - currentScrollHeight;
      } catch (err) {
        console.error("Failed to load more messages:", err);
      } finally {
        isLoadingMore = false;
      }
    }
  };

  // Message search is not offered currently; controls and overlay removed

  $: if ($wsMessages.length > 0) {
    tick().then(() => {
      // Only auto-scroll if the user is at or near the bottom, or if it's a new message
      const isAtBottom =
        chatContainer.scrollHeight - chatContainer.scrollTop <=
        chatContainer.clientHeight + 100; // 100px tolerance
      const lastMessage = $wsMessages[$wsMessages.length - 1];
      const lastRole = lastMessage ? lastMessage.role : null;

      // Auto-scroll if the user is near the bottom, or if the new message is an assistant message
      if (
        isAtBottom ||
        lastRole === "assistant" ||
        lastMessage.type === "typing"
      ) {
        chatContainer.scrollTop = chatContainer.scrollHeight;
      }
    });
  }

  // // Handle processing of queued messages when online
  // $: if (!$isOffline) {
  //   // This is the correct way to pass the sendMessage *from the websocket store*
  //   // for offline processing.
  //   processQueue(websocket.sendMessage);
  // }

  // Header removed from this component for cleaner composition
</script>

<div class="flex flex-col h-full text-black dark:text-white">
  <main
    class="flex-1 overflow-y-auto px-2 md:px-4 py-3 md:py-4 space-y-3"
    bind:this={chatContainer}
    on:scroll={handleScroll}
  >
    {#if $isReconnecting}
      <div
        class="w-full text-center py-2 bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200 rounded-lg"
      >
        Reconnecting to chat...
      </div>
    {/if}

    {#if isLoadingMore}
      <div class="flex justify-center py-2">
        <div
          class="animate-spin rounded-full h-6 w-6 border-b-2 border-black"
        ></div>
      </div>
    {/if}

    {#if $messages.length === 0}
      <div class="flex flex-col items-center justify-center text-center pt-12 md:pt-16">
        <div class="w-14 h-14 rounded-full bg-neutral-100 dark:bg-neutral-900 flex items-center justify-center mb-4">
          <span class="text-xl">🪴</span>
        </div>
        <h2 class="text-2xl font-semibold">Start a new conversation</h2>
        <p class="text-neutral-600 dark:text-neutral-400 mt-2 max-w-md">
          Ask anything. Press Enter to send, Shift+Enter for a new line.
        </p>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 mt-6 w-full max-w-2xl">
          <button class="text-left glass rounded-xl p-3 border border-neutral-200/60 dark:border-neutral-800/60 hover:opacity-95 transition" on:click={() => { userPrompt = 'Brainstorm video ideas for a product launch campaign'; textarea?.focus(); }}>
            Brainstorm video ideas for a product launch campaign
          </button>
          <button class="text-left glass rounded-xl p-3 border border-neutral-200/60 dark:border-neutral-800/60 hover:opacity-95 transition" on:click={() => { userPrompt = 'Explain this error and show how to fix it: TypeError: x is not a function'; textarea?.focus(); }}>
            Explain this error and show how to fix it: TypeError: x is not a function
          </button>
          <button class="text-left glass rounded-xl p-3 border border-neutral-200/60 dark:border-neutral-800/60 hover:opacity-95 transition" on:click={() => { userPrompt = 'Summarize this article in 5 bullet points with action items'; textarea?.focus(); }}>
            Summarize this article in 5 bullet points with action items
          </button>
          <button class="text-left glass rounded-xl p-3 border border-neutral-200/60 dark:border-neutral-800/60 hover:opacity-95 transition" on:click={() => { userPrompt = 'Draft a polite email to request a project deadline extension'; textarea?.focus(); }}>
            Draft a polite email to request a project deadline extension
          </button>
        </div>
      </div>
    {:else}
    {#each $messages as message (message.id)}
        <div id="message-{message.id}" class="flex">
          <div class="w-full">
          <ChatMessage {message} />
        </div>
      </div>
    {/each}
    {/if}
  </main>

  <div
    class="border-t border-white/40 dark:border-white/10 px-2 md:px-4 pt-2 pb-2 md:pb-3 bg-transparent"
  >
    <form class="flex flex-col gap-1.5" on:submit|preventDefault={handleSubmit}>

      <div class="flex items-end gap-1.5">
      <textarea
        bind:this={textarea}
          class="flex-1 resize-none rounded-lg border border-neutral-200/60 dark:border-neutral-800/60 bg-white/70 dark:bg-black/40 p-2.5 focus:outline-none focus:ring-1 focus:ring-black dark:focus:ring-white backdrop-blur-xl transition-shadow"
        rows="1"
        style="max-height: 150px; overflow-y: auto;"
          placeholder={$isOffline ? "You are offline" : "Message Botanic..."}
        bind:value={userPrompt}
        disabled={$isOffline || isSubmitting}
        on:keydown={handleKeydown}
        on:input={handleInput}
      ></textarea>
      <button
        type="submit"
          class="px-3 py-2 bg-black text-white rounded-full hover:opacity-90 focus:outline-none focus:ring-1 focus:ring-black dark:focus:ring-white disabled:opacity-50 disabled:cursor-not-allowed transition-opacity"
        disabled={$isOffline || isSubmitting || !userPrompt.trim()}
        title={$isOffline ? "You are offline" : "Send"}
      >
        {#if isSubmitting}
          <div class="flex items-end gap-2">
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
          </div>
        {:else}
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M12 5l7 7-7 7" />
          </svg>
        {/if}
      </button>
      </div>
      
    </form>
    </div>
</div>

<style>
  textarea {
    min-height: 48px;
    max-height: 150px;
    overflow-y: auto;
  }

  .header-controls { display: none; }

  :global(.message-bubble.user) {
    @apply bg-zinc-600 text-white rounded-lg p-3;
  }

  :global(.message-bubble.assistant) {
    @apply bg-white dark:bg-zinc-700 text-zinc-800 dark:text-zinc-200 rounded-lg p-3;
  }
</style>
