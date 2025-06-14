<script lang="ts">
  import { onMount } from "svelte";
  import { auth } from "$lib/stores/auth";
  import { goto } from "$app/navigation";
  import { api } from "$lib/api/client";
  import { llmStore } from "$lib/stores/llm";
  import ChatList from "$lib/components/ChatList.svelte";
  import ChatInterface from "$lib/components/ChatInterface.svelte";
  import { sessions, activeSessionId } from "$lib/stores";
  import { store as websocket } from "$lib/stores/websocket";
  import { isOffline } from '$lib/stores/offline';
  import { notifications } from '$lib/stores/notifications';
  import { ApiError } from '$lib/api/client';
  import type { Message } from '$lib/types';

  let isLoading = true;
  let showSidebar = true;

  onMount(async () => {
    try {
      const isAuthenticated = await auth.checkAuth();
      if (!isAuthenticated) {
        console.log('Not authenticated, redirecting to login...');
        goto("/");
        return;
      }

      await llmStore.loadModels();

      // Get URL parameters
      const urlParams = new URLSearchParams(window.location.search);
      const initialMessage = urlParams.get('message');
      const selectedModel = urlParams.get('model');

      // If we have an initial message, create a new session
      if (initialMessage && selectedModel) {
        try {
          // Create a new chat session
          const session = await api.createSession({
            title: initialMessage.slice(0, 50) + (initialMessage.length > 50 ? '...' : ''),
            model: selectedModel
          });

          if (!session || !session.id) {
            throw new Error('Failed to create session: Invalid response');
          }

          // Set this as the active session first
          $activeSessionId = session.id;
          
          // Add the session to the store
          $sessions = [{
            id: session.id,
            user_id: session.user_id,
            title: session.title,
            model: session.model || 'default',
            created_at: session.created_at,
            updated_at: session.updated_at,
            messages: []
          }, ...$sessions];

          try {
            // Send the initial message
            await api.sendMessage(session.id, initialMessage);
          } catch (error) {
            console.error('Failed to send initial message:', error);
            notifications.add({
              type: 'error',
              message: 'Failed to send initial message',
              duration: 5000
            });
          }

          // Clear URL parameters
          window.history.replaceState({}, '', '/chat');

        } catch (error) {
          console.error('Failed to create initial chat:', error);
          notifications.add({
            type: 'error',
            message: 'Failed to create chat session',
            duration: 5000
          });
          // Redirect to home on failure
          goto('/');
          return;
        }
      }

      try {
        console.log('Fetching chat sessions...');
        const userSessions = await api.getSessions();
        console.log('Received sessions:', userSessions);
        $sessions = userSessions.map(session => ({
          id: session.id,
          user_id: session.user_id,
          title: session.title,
          model: session.model || 'default',
          created_at: session.created_at,
          updated_at: session.updated_at,
          messages: (session.messages || []).map(msg => ({
            id: msg.id,
            session_id: msg.session_id,
            user_id: msg.user_id,
            content: msg.content,
            model: msg.model || 'default',
            type: msg.type || 'message',
            created_at: msg.created_at,
            updated_at: msg.updated_at || msg.created_at,
            role: msg.role || 'user'
          }))
        }));
        if (userSessions.length > 0 && !$activeSessionId) {
          $activeSessionId = userSessions[0].id;
        }
      } catch (error) {
        console.error("Failed to load sessions:", error);
        if (error instanceof ApiError && error.status === 401) {
          console.log("Session expired, redirecting to login...");
          await auth.clearAuth();
          goto("/");
        } else {
          notifications.add({
            type: 'error',
            message: 'Failed to load chat sessions',
            duration: 5000
          });
        }
      }
    } catch (error) {
      console.error("Failed to initialize:", error);
      notifications.add({
        type: 'error',
        message: 'Failed to initialize app',
        duration: 5000
      });
    } finally {
      isLoading = false;
    }
  });

  function toggleSidebar() {
    showSidebar = !showSidebar;
  }
</script>

<div class="flex h-[calc(100vh-4rem)] bg-white dark:bg-black">
  {#if isLoading}
    <!-- Skeleton Loader -->
    <div class="flex w-full h-full animate-pulse">
      <!-- Skeleton Sidebar -->
      <div class="w-80 border-r border-neutral-200 dark:border-neutral-800 p-4 space-y-4 hidden md:block">
        <div class="h-10 bg-neutral-200 dark:bg-neutral-800 rounded-md"></div>
        <div class="space-y-3 pt-4">
          <div class="h-8 bg-neutral-200 dark:bg-neutral-800 rounded-md"></div>
          <div class="h-8 bg-neutral-200 dark:bg-neutral-800 rounded-md w-5/6"></div>
          <div class="h-8 bg-neutral-200 dark:bg-neutral-800 rounded-md w-3/4"></div>
          <div class="h-8 bg-neutral-200 dark:bg-neutral-800 rounded-md w-5/6"></div>
          <div class="h-8 bg-neutral-200 dark:bg-neutral-800 rounded-md"></div>
          <div class="h-8 bg-neutral-200 dark:bg-neutral-800 rounded-md w-1/2"></div>
        </div>
      </div>

      <!-- Skeleton Main Chat Area -->
      <div class="flex-1 flex flex-col">
        <!-- Skeleton Header -->
        <div class="flex items-center justify-between p-6 border-b border-neutral-200 dark:border-neutral-800">
          <div class="flex items-center gap-4">
            <div class="w-6 h-6 bg-neutral-200 dark:bg-neutral-800 rounded"></div>
            <div class="h-6 w-48 bg-neutral-200 dark:bg-neutral-800 rounded"></div>
          </div>
        </div>

        <!-- Skeleton Chat Messages -->
        <div class="flex-1 p-6 space-y-6 overflow-y-auto">
          <!-- Bot Message Skeleton -->
          <div class="flex items-start gap-3">
            <div class="w-8 h-8 bg-neutral-200 dark:bg-neutral-800 rounded-full"></div>
            <div class="flex-1 space-y-2">
              <div class="h-4 w-3/4 bg-neutral-200 dark:bg-neutral-800 rounded"></div>
              <div class="h-4 w-1/2 bg-neutral-200 dark:bg-neutral-800 rounded"></div>
            </div>
          </div>
          <!-- User Message Skeleton -->
          <div class="flex items-start gap-3 justify-end">
            <div class="flex-1 space-y-2 text-right">
              <div class="h-4 w-3/4 bg-neutral-200 dark:bg-neutral-800 rounded ml-auto"></div>
            </div>
             <div class="w-8 h-8 bg-neutral-200 dark:bg-neutral-800 rounded-full"></div>
          </div>
          <!-- Bot Message Skeleton -->
          <div class="flex items-start gap-3">
            <div class="w-8 h-8 bg-neutral-200 dark:bg-neutral-800 rounded-full"></div>
            <div class="flex-1 space-y-2">
              <div class="h-4 w-full bg-neutral-200 dark:bg-neutral-800 rounded"></div>
              <div class="h-4 w-2/3 bg-neutral-200 dark:bg-neutral-800 rounded"></div>
            </div>
          </div>
        </div>

        <!-- Skeleton Message Input -->
        <div class="p-4 mb-14 mt-auto border-t border-neutral-200 dark:border-neutral-800">
           <div class="h-12 w-full bg-neutral-200 dark:bg-neutral-800 rounded-lg"></div>
        </div>
      </div>
    </div>
  {:else}
    <!-- Actual Content -->
    <div class="flex w-full">
      <!-- Sidebar -->
      <div 
        class="transition-all duration-300 ease-in-out {showSidebar ? 'w-80' : 'w-0'} overflow-hidden"
      >
        <!-- This inner div has a fixed width to prevent content from being squashed during transition -->
        <div class="h-full w-80 border-r border-neutral-200 dark:border-neutral-800">
          <ChatList />
        </div>
      </div>

      <!-- Main Chat Area -->
      <div class="flex-1 flex flex-col">
        <!-- Header -->
        <div class="flex items-center justify-between p-4 border-b border-neutral-200 dark:border-neutral-800">
          <div class="flex items-center gap-4">
            <button
              class="p-2 text-neutral-600 dark:text-neutral-400 hover:text-black dark:hover:text-white"
              on:click={toggleSidebar}
              title={showSidebar ? "Hide sidebar" : "Show sidebar"}
              aria-label="Toggle sidebar"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                class="w-6 h-6"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M4 6h16M4 12h16M4 18h16"
                />
              </svg>
            </button>
            <h1 class="text-xl font-semibold text-black dark:text-white">
              {#if $activeSessionId}
                {$sessions.find(s => s.id === $activeSessionId)?.title || 'New Chat'}
              {:else}
                New Chat
              {/if}
            </h1>
          </div>
          <div class="flex items-center gap-4">
            {#if $isOffline}
              <div class="flex items-center gap-2 text-yellow-600 dark:text-yellow-400">
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
                    d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                  />
                </svg>
                <span class="text-sm">Offline</span>
              </div>
            {/if}
          </div>
        </div>

        <!-- Chat Interface -->
        <div class="flex-1 overflow-hidden">
          {#if $activeSessionId && $activeSessionId !== 'undefined'}
            <ChatInterface sessionId={$activeSessionId} />
          {:else}
            <div class="flex flex-col items-center justify-center h-full gap-4 p-8 text-center">
              <div class="w-16 h-16 text-neutral-400 dark:text-neutral-600">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                  />
                </svg>
              </div>
              <h2 class="text-2xl font-semibold text-black dark:text-white">
                Welcome to Botanic Chat
              </h2>
              <p class="max-w-md text-neutral-600 dark:text-neutral-400">
                Select a chat from the sidebar or start a new conversation to begin.
              </p>
              <button
                class="px-6 py-3 mt-4 text-white bg-black dark:text-black dark:bg-white rounded-lg hover:bg-black/80 focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2"
                on:click={() => goto('/')}
              >
                Start New Chat
              </button>
            </div>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  :global(.markdown-content) {
    max-width: none;
  }

  :global(.markdown-content pre) {
    background-color: rgb(31, 41, 55);
    color: rgb(243, 244, 246);
    padding: 1rem;
    border-radius: 0.5rem;
    overflow-x: auto;
  }

  :global(.markdown-content code) {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
    font-size: 0.875rem;
  }

  :global(.markdown-content p) {
    margin-top: 0.5rem;
    margin-bottom: 0.5rem;
  }

  :global(.markdown-content ul) {
    list-style-type: disc;
    list-style-position: inside;
    margin-top: 0.5rem;
    margin-bottom: 0.5rem;
  }

  :global(.markdown-content ol) {
    list-style-type: decimal;
    list-style-position: inside;
    margin-top: 0.5rem;
    margin-bottom: 0.5rem;
  }

  :global(.markdown-content blockquote) {
    border-left-width: 4px;
    border-left-color: rgb(209, 213, 219);
    padding-left: 1rem;
    font-style: italic;
    margin-top: 0.5rem;
    margin-bottom: 0.5rem;
  }

  :global(.markdown-content table) {
    border-collapse: collapse;
    width: 100%;
    margin-top: 0.5rem;
    margin-bottom: 0.5rem;
  }

  :global(.markdown-content th),
  :global(.markdown-content td) {
    border: 1px solid rgb(209, 213, 219);
    padding: 0.5rem;
  }

  :global(.dark .markdown-content blockquote) {
    border-left-color: rgb(75, 85, 99);
  }

  :global(.dark .markdown-content th),
  :global(.dark .markdown-content td) {
    border-color: rgb(75, 85, 99);
  }

  :global(.dark .markdown-content th) {
    background-color: rgb(31, 41, 55);
  }
</style>
