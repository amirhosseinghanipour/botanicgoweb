<script lang="ts">
  import { onMount } from 'svelte';
  import { llmStore } from '$lib/stores/llm';
  import type { Model } from '$lib/stores/llm';
  import { icons } from '$lib/icons.js';
  import { goto } from '$app/navigation';
  import { auth } from '$lib/stores/auth';
  import { api } from '$lib/api/client';
  import { notifications } from '$lib/stores/notifications';

  const DEFAULT_MODEL_ID = "deepseek/deepseek-chat:free";
  const SELECTED_MODEL_KEY = "lastSelectedModelId";

  let observer: IntersectionObserver;
  let loadingTrigger: HTMLElement;
  let promptTextValue = '';
  let selectedModelId = DEFAULT_MODEL_ID;
  let searchQuery = '';
  let isDropdownOpen = false;
  let dropdownRef: HTMLDivElement;
  let isSubmitting = false;

  // Top free models from famous providers
  const TOP_FREE_MODELS = [
    "deepseek/deepseek-chat:free",
    "mistralai/Mistral-7B-Instruct-v0.2",
    "google/gemma-2b-it",
    "meta-llama/Llama-2-7b-chat-hf",
    "microsoft/phi-2"
  ];

  $: filteredModels = searchQuery === ""
    ? TOP_FREE_MODELS.map(id => [...$llmStore.freeModels, ...$llmStore.nonFreeModels].find(m => m.id === id)).filter((model): model is Model => model !== undefined)
    : [...$llmStore.freeModels, ...$llmStore.nonFreeModels].filter(model =>
        model.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        model.description.toLowerCase().includes(searchQuery.toLowerCase())
      );

  function updatePrompt(value: string) {
    promptTextValue = value;
  }

  async function handleSubmit() {
    if (!promptTextValue.trim() || isSubmitting) return;
    
    if (!$auth.user) {
      await goto('/login');
      return;
    }

    isSubmitting = true;
    try {
      // Encode the message and model as URL parameters
      const params = new URLSearchParams({
        message: promptTextValue,
        model: selectedModelId
      });

      // Redirect to chat with the parameters
      goto(`/chat?${params.toString()}`);
    } catch (error) {
      console.error('Failed to start chat:', error);
      notifications.add({
        type: 'error',
        message: 'Failed to start chat',
        duration: 5000
      });
    } finally {
      isSubmitting = false;
    }
  }

  function handleModelSelect(model: Model) {
    selectedModelId = model.id;
    llmStore.selectModel(model);
    localStorage.setItem(SELECTED_MODEL_KEY, model.id);
  }

  function handleClickOutside(event: MouseEvent) {
    if (dropdownRef && !dropdownRef.contains(event.target as Node)) {
      isDropdownOpen = false;
    }
  }

  onMount(() => {
    // Load initial models
    llmStore.loadModels();

    // Load last selected model
    const saved = localStorage.getItem(SELECTED_MODEL_KEY);
    if (saved) {
      const model = [...$llmStore.freeModels, ...$llmStore.nonFreeModels].find(m => m.id === saved);
      if (model) {
        selectedModelId = model.id;
        llmStore.selectModel(model);
      }
    }

    // Set up intersection observer for infinite scroll
    observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) {
          llmStore.loadMoreModels();
        }
      },
      { threshold: 0.1 }
    );

    if (loadingTrigger) {
      observer.observe(loadingTrigger);
    }

    document.addEventListener('click', handleClickOutside);
    return () => {
      if (observer) {
        observer.disconnect();
      }
      document.removeEventListener('click', handleClickOutside);
    };
  });

  function formatPrice(price: string): string {
    const num = parseFloat(price);
    if (isNaN(num)) return 'N/A';
    return num === 0 ? 'Free' : `$${num.toFixed(4)}/token`;
  }
</script>

<div class="flex flex-col items-center text-center max-w-7xl mx-auto px-4">
  <h1 class="text-5xl md:text-6xl font-bold tracking-tight mt-12 mb-4 text-black dark:text-white">
    Ask Botanic Anything
  </h1>
  <p class="max-w-2xl text-base text-neutral-600 dark:text-neutral-400 mb-10">
    An open-source AI chat interface. Built with love under war in the middle-east.
  </p>

  <div class="w-full max-w-4xl mb-12">
    <div class="relative w-full">
      <textarea
        bind:value={promptTextValue}
        on:input={(e) => updatePrompt((e.target as HTMLTextAreaElement).value)}
        on:keydown={(e) => {
          if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSubmit();
          }
        }}
        rows="5"
        placeholder="Ask anything or describe what you want to create..."
        class="w-full pl-6 pr-20 py-5 bg-white dark:bg-black rounded-2xl text-lg outline-none resize-none align-bottom placeholder-neutral-500 dark:placeholder-neutral-600 border border-neutral-200 dark:border-neutral-800 focus:ring-2 focus:ring-black dark:focus:ring-white transition-shadow duration-300"
        aria-label="Enter your prompt"
      ></textarea>
      <button
        on:click={handleSubmit}
        disabled={!promptTextValue.trim() || isSubmitting}
        class="absolute top-5 right-5 w-10 h-10 flex items-center justify-center bg-black dark:bg-white text-white dark:text-black font-bold rounded-full disabled:opacity-30 disabled:cursor-not-allowed transition-all duration-300 hover:scale-105"
        aria-label="Submit prompt"
      >
        {#if isSubmitting}
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
            <span>Creating...</span>
          </div>
        {:else}
          Send Message
        {/if}
      </button>
    </div>
    <div class="mt-4 flex items-center gap-3">
      <button
        class="flex items-center justify-center w-10 h-10 rounded-full bg-neutral-100 dark:bg-neutral-900 text-black dark:text-white hover:bg-neutral-200 dark:hover:bg-neutral-800 transition-colors"
        aria-label="Upload file"
      >
        {@html icons.upload}
      </button>
      <button
        class="flex items-center justify-center w-10 h-10 rounded-full bg-neutral-100 dark:bg-neutral-900 text-black dark:text-white hover:bg-neutral-200 dark:hover:bg-neutral-800 transition-colors"
        aria-label="Think mode"
      >
        {@html icons.think}
      </button>
      <div class="flex-grow"></div>
      {#if !$llmStore.isLoading && !$llmStore.error && ($llmStore.freeModels.length > 0 || $llmStore.nonFreeModels.length > 0)}
        <div class="relative" bind:this={dropdownRef}>
          <button
            on:click={() => isDropdownOpen = !isDropdownOpen}
            class="flex items-center gap-2 px-4 py-2.5 text-sm font-medium bg-white dark:bg-black border border-neutral-200 dark:border-neutral-800 rounded-lg hover:border-black dark:hover:border-white transition-colors"
          >
            <span>{[...$llmStore.freeModels, ...$llmStore.nonFreeModels].find(m => m.id === selectedModelId)?.name || 'Select Model'}</span>
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </button>
          
          {#if isDropdownOpen}
            <div class="absolute right-0 mt-2 w-72 bg-white dark:bg-black border border-neutral-200 dark:border-neutral-800 rounded-lg shadow-lg z-50">
              <div class="p-2 border-b border-neutral-200 dark:border-neutral-800">
                <input
                  type="text"
                  bind:value={searchQuery}
                  placeholder="Search models..."
                  class="w-full px-3 py-2 text-sm bg-neutral-100 dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md focus:outline-none focus:ring-2 focus:ring-black dark:focus:ring-white"
                />
              </div>
              <div class="max-h-60 overflow-y-auto">
                {#each filteredModels as model}
                  <button
                    on:click={() => {
                      handleModelSelect(model);
                      isDropdownOpen = false;
                    }}
                    class="w-full px-4 py-2 text-left hover:bg-neutral-100 dark:hover:bg-neutral-900 {model.id === selectedModelId ? 'bg-neutral-100 dark:bg-neutral-900' : ''}"
                  >
                    <div class="font-medium">{model.name}</div>
                    <div class="text-xs text-neutral-500 dark:text-neutral-400 truncate">
                      {model.description}
                    </div>
                  </button>
                {/each}
              </div>
            </div>
          {/if}
        </div>
      {/if}
    </div>
  </div>

  <div class="w-full">
    {#if $llmStore.isLoading}
      <div class="w-full grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5 animate-pulse">
        {#each Array(8) as _}
          <div class="p-5 border border-neutral-200 dark:border-neutral-800 rounded-xl bg-white dark:bg-black">
            <div class="w-8 h-8 mb-4 bg-neutral-100 dark:bg-neutral-900 rounded-md"></div>
            <div class="w-3/4 h-4 mb-2 bg-neutral-100 dark:bg-neutral-900 rounded-md"></div>
            <div class="w-full h-3 bg-neutral-100 dark:bg-neutral-900 rounded-md"></div>
            <div class="w-1/2 h-3 mt-1.5 bg-neutral-100 dark:bg-neutral-900 rounded-md"></div>
          </div>
        {/each}
      </div>
    {:else if $llmStore.error}
      <div class="w-full text-center p-10 bg-neutral-50 dark:bg-neutral-900/50 border border-neutral-200 dark:border-neutral-800 rounded-lg">
        <h3 class="font-semibold text-black dark:text-white">Failed to Load Models</h3>
        <p class="text-sm text-neutral-600 dark:text-neutral-400 mt-2 max-w-md mx-auto">
          There was a problem fetching data from the API. This can happen if the service is down or there is a network issue.
        </p>
        <p class="text-xs text-neutral-500 dark:text-neutral-600 mt-4">
          <strong>Error:</strong>
          {$llmStore.error}
        </p>
      </div>
    {:else if $llmStore.freeModels.length === 0 && $llmStore.nonFreeModels.length === 0}
      <div class="text-center text-gray-500 py-8">
        No models available at the moment.
      </div>
    {:else}
      <!-- Free Models Section -->
      {#if $llmStore.freeModels.length > 0}
        <div class="mb-8">
          <h2 class="text-2xl font-semibold mb-4">Free Models</h2>
          <div class="w-full grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5">
            {#each $llmStore.freeModels as model}
              <button
                on:click={() => handleModelSelect(model)}
                class="relative text-left p-5 border rounded-xl transition-all duration-300 ease-in-out transform hover:-translate-y-1 {selectedModelId === model.id ? 'bg-black dark:bg-white text-white dark:text-black border-black dark:border-white' : 'bg-white dark:bg-black border-neutral-200 dark:border-neutral-800 hover:border-black dark:hover:border-white'}"
                aria-label={`Select ${model.name}`}
                role="button"
              >
                <div class="relative">
                  <div class="w-8 h-8 mb-4">
                    {@html icons.model}
                  </div>
                  <h3 class="font-semibold text-base mb-1.5">{model.name}</h3>
                  <p class="text-sm leading-relaxed line-clamp-2 {selectedModelId === model.id ? 'text-neutral-400 dark:text-neutral-500' : 'text-neutral-600 dark:text-neutral-400'}">
                    {model.description}
                  </p>
                  <div class="text-xs mt-2">
                    <span>Context: {model.context_length}</span> &middot; <span>Prompt: {formatPrice(model.pricing.prompt)}</span> &middot; <span>Completion: {formatPrice(model.pricing.completion)}</span>
                  </div>
                </div>
              </button>
            {/each}
          </div>
        </div>
      {/if}
      <!-- Divider with Text -->
      {#if $llmStore.freeModels.length > 0 && $llmStore.nonFreeModels.length > 0}
        <div class="relative my-8">
          <div class="absolute inset-0 flex items-center">
            <div class="w-full border-t border-gray-300"></div>
          </div>
          <div class="relative flex justify-center">
            <span class="bg-white px-4 text-gray-500">Premium Models</span>
          </div>
        </div>
      {/if}
      <!-- Non-Free Models Section -->
      {#if $llmStore.nonFreeModels.length > 0}
        <div class="mb-8">
          <div class="w-full grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5">
            {#each $llmStore.nonFreeModels as model}
              <button
                on:click={() => handleModelSelect(model)}
                class="relative text-left p-5 border rounded-xl transition-all duration-300 ease-in-out transform hover:-translate-y-1 {selectedModelId === model.id ? 'bg-black dark:bg-white text-white dark:text-black border-black dark:border-white' : 'bg-white dark:bg-black border-neutral-200 dark:border-neutral-800 hover:border-black dark:hover:border-white'}"
                aria-label={`Select ${model.name}`}
                role="button"
              >
                <div class="relative">
                  <div class="w-8 h-8 mb-4">
                    {@html icons.model}
                  </div>
                  <h3 class="font-semibold text-base mb-1.5">{model.name}</h3>
                  <p class="text-sm leading-relaxed line-clamp-2 {selectedModelId === model.id ? 'text-neutral-400 dark:text-neutral-500' : 'text-neutral-600 dark:text-neutral-400'}">
                    {model.description}
                  </p>
                  <div class="text-xs mt-2">
                    <span>Context: {model.context_length}</span> &middot; <span>Prompt: {formatPrice(model.pricing.prompt)}</span> &middot; <span>Completion: {formatPrice(model.pricing.completion)}</span>
                  </div>
                </div>
              </button>
            {/each}
          </div>
        </div>
      {/if}
      <!-- Loading Trigger for Infinite Scroll -->
      {#if $llmStore.hasMore}
        <div bind:this={loadingTrigger} class="h-10 flex items-center justify-center">
          {#if $llmStore.isLoading}
            <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-500"></div>
          {/if}
        </div>
      {/if}
    {/if}
  </div>
</div>
