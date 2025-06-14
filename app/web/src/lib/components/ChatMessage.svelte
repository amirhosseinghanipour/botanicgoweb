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

  $: modelName = $llmStore.models.find(m => m.id === message.model)?.name || message.model;
  $: isUser = browser && message.type === 'message' && message.userId === localStorage.getItem('userId');

  const handleCopy = () => {
    if (browser) {
      navigator.clipboard.writeText(message.content);
    }
  };

  const formatTimestamp = () => {
    return message.createdAt
      ? new Date(message.createdAt).toLocaleTimeString([], {
          hour: "2-digit",
          minute: "2-digit",
        })
      : "";
  };
</script>

<div
  class="w-full flex {isUser ? 'justify-end' : 'justify-start'}"
>
  <div
    class="max-w-xl lg:max-w-3xl flex flex-col px-4 py-2 my-1 relative group"
  >
    <div
      class="text-base leading-relaxed break-words {isUser
        ? 'border border-gray-200 rounded-l-3xl rounded-tr-3xl p-3 text-black dark:text-white'
        : 'text-neutral-800 dark:text-neutral-200'}"
    >
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
    <div
      class={isUser
        ? "flex gap-2 mt-2 opacity-0 group-hover:opacity-100 transition-opacity"
        : "flex gap-2 mt-2"}
    >
      {#if message.type === 'message'}
        <button
          on:click={handleCopy}
          class="p-1 rounded hover:bg-neutral-200 dark:hover:bg-neutral-800"
          title="Copy"
        >
          {@html icons.copy}
        </button>
        <span class="text-xs text-neutral-500 self-center"
          >{formatTimestamp()}</span
        >
      {/if}
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
