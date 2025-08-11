<script lang="ts">
  import { icons } from "$lib/icons.js";
  import { renderMarkdown } from '$lib/utils/markdown';
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import type { Message } from '$lib/types';
  import { llmStore } from '$lib/stores/llm';

  export let message: Message;

  let renderedContent = '';
  $: if (message.type === 'message') {
    renderedContent = renderMarkdown(message.content);
  }

  // Support both camelCase and snake_case from different sources
  $: messageUserId = (message as any).userId ?? (message as any).user_id ?? '';
  $: messageCreatedAt = (message as any).createdAt ?? (message as any).created_at ?? '';
  $: modelName = $llmStore.models.find(m => m.id === message.model)?.name || message.model;
  $: isUser = browser && message.type === 'message' && messageUserId === localStorage.getItem('userId');

  const handleCopy = () => {
    if (browser) {
      navigator.clipboard.writeText(message.content);
    }
  };

  const formatTimestamp = () => {
    return messageCreatedAt
      ? new Date(messageCreatedAt).toLocaleTimeString([], {
          hour: "2-digit",
          minute: "2-digit",
        })
      : "";
  };
</script>

<div class="w-full">
  <div class="flex items-start gap-3 px-2 py-2 my-0.5 {isUser ? 'justify-end flex-row-reverse' : ''}">
    
    <div class="shrink-0 w-8 h-8 rounded-full flex items-center justify-center {isUser ? 'bg-neutral-800 text-white dark:bg-white dark:text-black' : 'bg-neutral-100 text-neutral-700 dark:bg-neutral-900 dark:text-neutral-200'}">
      {#if isUser}
        <span class="text-sm">You</span>
      {:else}
        <span class="text-sm">AI</span>
      {/if}
    </div>
    <div class="max-w-[min(70ch,calc(100%-4rem))] flex-1 relative group">
      <div class="text-base leading-relaxed break-words {isUser
        ? 'bg-white/70 dark:bg-black/40 border border-neutral-200/70 dark:border-neutral-800/70 rounded-2xl p-3 text-black dark:text-white backdrop-blur-xl shadow-sm'
        : 'bg-white/60 dark:bg-neutral-900/50 border border-neutral-200/60 dark:border-neutral-800/60 rounded-2xl p-3 text-neutral-900 dark:text-neutral-100 backdrop-blur-xl shadow-sm'}">
      {#if message.type === 'typing'}
        <div class="flex items-center gap-1.5">
          <span class="w-2 h-2 bg-neutral-400 rounded-full animate-pulse"></span>
          <span
            class="w-2 h-2 bg-neutral-400 rounded-full animate-pulse"
            style="animation-delay: 0.2s;"
          ></span>
          <span
            class="w-2 h-2 bg-neutral-400 rounded-full animate-pulse"
            style="animation-delay: 0.4s;"
          ></span>
        </div>
      {:else if message.type === 'error'}
        <div class="text-sm text-red-500 italic">
          {message.content}
        </div>
      {:else if message.type === 'status'}
        <div class="text-sm text-neutral-500 italic">
          {message.content}
        </div>
      {:else}
        <div class="markdown-content">
          {@html renderedContent}
        </div>
      {/if}
      </div>
      <div class="flex items-center gap-2 mt-1 {isUser ? 'justify-end' : ''}">
        {#if message.type === 'message'}
          <button
            on:click={handleCopy}
            class="p-1 rounded hover:bg-neutral-200 dark:hover:bg-neutral-800"
            title="Copy"
          >
            {@html icons.copy}
          </button>
          <span class="text-[11px] text-neutral-500">
            {formatTimestamp()}
          </span>
        {/if}
      </div>
    </div>
  </div>
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

  :global(.markdown-content th) {
    background-color: rgb(243, 244, 246);
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
