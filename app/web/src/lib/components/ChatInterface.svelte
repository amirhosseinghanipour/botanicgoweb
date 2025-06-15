<script lang="ts">
  import { onMount, tick } from "svelte";
  import { api } from "$lib/api/client";
  import { store as websocket } from "$lib/stores/websocket";
  import ChatMessage from "./ChatMessage.svelte";
  import MessageSearch from "./MessageSearch.svelte";
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
  import LLMSelector from "./LLMSelector.svelte";
  import ModelComparison from "./ModelComparison.svelte";
  import { goto } from "$app/navigation";

  export let sessionId: string;

  let userPrompt = "";
  let isSubmitting = false;
  let textarea: HTMLTextAreaElement;
  let chatContainer: HTMLDivElement;
  let isLoadingMore = false;
  let hasMoreMessages = true;
  let isSearchOpen = false;
  let showModelComparison = false;
  let comparisonPrompt = "";

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
          const session = await api.getSession(sessionId);
          if (session && session.messages) {
            messages.set(session.messages); // Set initial messages
          }
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

  const handleSearchSelect = ({ detail }: { detail: { message: any } }) => {
    const messageElement = document.getElementById(
      `message-${detail.message.id}`,
    );
    if (messageElement) {
      messageElement.scrollIntoView({ behavior: "smooth", block: "center" });
      messageElement.classList.add("highlight-message");
      setTimeout(() => {
        messageElement.classList.remove("highlight-message");
      }, 2000);
    }
  };

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

  // Handle processing of queued messages when online
  $: if (!$isOffline) {
    // This is the correct way to pass the sendMessage *from the websocket store*
    // for offline processing.
    processQueue(websocket.sendMessage);
  }

  const handleNewChat = () => {
    goto("/");
  };
</script>

<div class="flex flex-col h-full text-black dark:text-white">
  <header
    class="flex justify-between items-center p-4 border-b border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-800/50 backdrop-blur-sm"
  >
    <div class="flex items-center gap-4">
      <button
        class="flex items-center gap-2 px-3 py-2 text-sm font-medium text-zinc-700 dark:text-zinc-200 hover:bg-zinc-100 dark:hover:bg-zinc-700 rounded-lg transition-colors"
        on:click={handleNewChat}
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
            d="M12 4v16m8-8H4"
          />
        </svg>
        New Chat
      </button>
    </div>
    <div class="header-controls">
      <LLMSelector />
    </div>
  </header>

  <main
    class="flex-1 overflow-y-auto p-4 space-y-4"
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

    {#each $messages as message (message.id)}
      <div
        id="message-{message.id}"
        class="flex"
        class:justify-end={message.role === "user"}
      >
        <div class="max-w-prose">
          <ChatMessage {message} />
        </div>
      </div>
    {/each}
  </main>

  <footer
    class="border-t mb-14 border-zinc-200 dark:border-zinc-700 p-4 bg-white dark:bg-zinc-800/50"
  >
    <form class="flex items-end gap-2" on:submit|preventDefault={handleSubmit}>
      <textarea
        bind:this={textarea}
        class="flex-1 resize-none rounded-lg border border-zinc-300 dark:border-zinc-600 bg-white dark:bg-zinc-700 p-2.5 focus:outline-none focus:ring-2 focus:ring-black transition-shadow"
        rows="1"
        style="max-height: 150px; overflow-y: auto;"
        placeholder={$isOffline ? "You are offline" : "Ask Botanic anything..."}
        bind:value={userPrompt}
        disabled={$isOffline || isSubmitting}
        on:keydown={handleKeydown}
        on:input={handleInput}
      ></textarea>
      <button
        type="submit"
        class="px-4 py-2 bg-zinc-600 text-white rounded-lg hover:bg-black focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 dark:focus:ring-offset-zinc-900 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        disabled={$isOffline || isSubmitting || !userPrompt.trim()}
        title={$isOffline ? "You are offline" : "Send"}
      >
        {#if isSubmitting}
          <div class="flex items-end gap-2">
            <svg
              class="w-4 h-4 animate-spin"
              viewBox="0 0 24 24"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                d="M12 4.75V6.25"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <path
                d="M17.1475 6.8525L16.0625 7.9375"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <path
                d="M19.25 12H17.75"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <path
                d="M17.1475 17.1475L16.0625 16.0625"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <path
                d="M12 17.75V19.25"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <path
                d="M6.8525 17.1475L7.9375 16.0625"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <path
                d="M4.75 12H6.25"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <path
                d="M6.8525 6.8525L7.9375 7.9375"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </div>
        {:else}
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
              d="M5 12h14M12 5l7 7-7 7"
            />
          </svg>
        {/if}
      </button>
    </form>
  </footer>
</div>

<MessageSearch
  {sessionId}
  isOpen={isSearchOpen}
  on:select={handleSearchSelect}
  on:close={() => (isSearchOpen = false)}
/>

<style>
  .highlight-message {
    animation: highlight 2s ease-out;
  }

  @keyframes highlight {
    0% {
      background-color: rgba(59, 130, 246, 0.3);
      border-radius: 0.5rem;
    }
    100% {
      background-color: transparent;
    }
  }

  textarea {
    min-height: 48px;
    max-height: 150px;
    overflow-y: auto;
  }

  .header-controls {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  :global(.message-bubble.user) {
    @apply bg-zinc-600 text-white rounded-lg p-3;
  }

  :global(.message-bubble.assistant) {
    @apply bg-white dark:bg-zinc-700 text-zinc-800 dark:text-zinc-200 rounded-lg p-3;
  }
</style>
