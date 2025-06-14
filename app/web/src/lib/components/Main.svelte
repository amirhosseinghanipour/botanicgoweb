<script lang="ts">
  import { onMount, createEventDispatcher } from "svelte";
  import {
    models,
    isLoading,
    error,
    selectedModelId,
    promptText,
    type Model
  } from "$lib/stores";
  import { icons } from "$lib/icons.js";
  import { goto } from '$app/navigation';
  import { api } from '$lib/api/client';
  import { notifications } from '$lib/stores/notifications';
  import { auth } from '$lib/stores/auth';

  const DEFAULT_MODEL_ID = "deepseek/deepseek-chat:free";
  const SELECTED_MODEL_KEY = "lastSelectedModelId";
  const CACHE_KEY = "cachedModels";
  const CACHE_TIMESTAMP_KEY = "modelsCacheTimestamp";
  const CACHE_DURATION_MS = 24 * 60 * 60 * 1000;

  let searchQuery = "";
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
    ? TOP_FREE_MODELS.map(id => $models.find(m => m.id === id)).filter((model): model is Model => model !== undefined)
    : $models.filter(model =>
        model.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        model.description.toLowerCase().includes(searchQuery.toLowerCase())
      );

  const loadLastSelectedModel = () => {
    const saved = localStorage.getItem(SELECTED_MODEL_KEY);
    if (saved && $models.find(m => m.id === saved)) {
      selectedModelId.set(saved);
    } else {
      selectedModelId.set(DEFAULT_MODEL_ID);
    }
  };

  const saveSelectedModel = (id: string) => {
    localStorage.setItem(SELECTED_MODEL_KEY, id);
    selectedModelId.set(id);
  };

  onMount(async () => {
    const loadFromCache = () => {
      const cachedData = localStorage.getItem(CACHE_KEY);
      if (cachedData) {
        try {
          const parsedData = JSON.parse(cachedData) as Model[];
          models.set(parsedData);
          loadLastSelectedModel();
          return true;
        } catch (e) {
          console.warn("Failed to parse cache:", e);
          return false;
        }
      }
      return false;
    };

    const refreshModels = async () => {
      try {
        const response = await fetch("https://openrouter.ai/api/v1/models");
        if (!response.ok) {
          throw new Error(`API request failed: ${response.status} ${response.statusText}`);
        }

        const data = await response.json() as { data: Model[] };
        const freeModels = data.data.filter(
          model => model.pricing.prompt === "0" && model.pricing.completion === "0"
        );

        if (freeModels.length === 0) {
          throw new Error("API returned no free models.");
        }

        const currentModels = $models;
        const hasChanges = JSON.stringify(freeModels) !== JSON.stringify(currentModels);
        
        if (hasChanges) {
          models.set(freeModels);
          localStorage.setItem(CACHE_KEY, JSON.stringify(freeModels));
          localStorage.setItem(CACHE_TIMESTAMP_KEY, Date.now().toString());
        }

        loadLastSelectedModel();
        return true;
      } catch (e) {
        console.error("Could not fetch models:", e);
        error.set(e instanceof Error ? e.message : "Failed to fetch models");
        return false;
      }
    };

    isLoading.set(true);
    const cacheLoaded = loadFromCache();

    if (!cacheLoaded) {
      await refreshModels();
    } else {
      const cachedTimestamp = localStorage.getItem(CACHE_TIMESTAMP_KEY);
      const isCacheStale = !cachedTimestamp || 
        Date.now() - parseInt(cachedTimestamp) > CACHE_DURATION_MS;
      if (isCacheStale) {
        refreshModels();
      }
    }

    isLoading.set(false);
  });

  const dispatch = createEventDispatcher<{
    submit: { prompt: string; modelId: string };
  }>();

  // Remove old handleSubmit, use handleSend instead
  async function handleSend() {
    if (!($promptText || '').trim() || isSubmitting) return;

    if (!$auth.user) {
      await goto('/login');
      return;
    }

    isSubmitting = true;
    try {
      // Create a new chat session
      const session = await api.createSession({
        title: $promptText.slice(0, 50) + ($promptText.length > 50 ? '...' : ''),
        model: $selectedModelId || DEFAULT_MODEL_ID
      });

      if (!session || !session.id) {
        throw new Error('Failed to create session: Invalid response');
      }

      // Send the initial message
      await api.sendMessage(session.id, $promptText);

      // Clear the prompt text
      promptText.set('');

      // Redirect to the chat with the session ID
      goto(`/chat/${session.id}`);
    } catch (error) {
      console.error('Failed to create chat:', error);
      notifications.add({
        type: 'error',
        message: 'Failed to create chat session',
        duration: 5000
      });
    } finally {
      isSubmitting = false;
    }
  }

  const debounce = <T extends (...args: any[]) => void>(fn: T, ms: number) => {
    let timeout: ReturnType<typeof setTimeout> | undefined;
    return (...args: Parameters<T>) => {
      if (timeout) clearTimeout(timeout);
      timeout = setTimeout(() => fn(...args), ms);
    };
  };

  const updatePrompt = debounce((value: string) => {
    promptText.set(value);
  }, 300);

  function handleClickOutside(event: MouseEvent) {
    if (dropdownRef && !dropdownRef.contains(event.target as Node)) {
      isDropdownOpen = false;
    }
  }

  onMount(() => {
    document.addEventListener('click', handleClickOutside);
    return () => {
      document.removeEventListener('click', handleClickOutside);
    };
  });
</script>


<div class="flex flex-col items-center text-center max-w-7xl mx-auto px-4">
  <h1
    class="text-5xl md:text-6xl font-bold tracking-tight mt-12 mb-4 text-black dark:text-white"
  >
    Ask Botanic Anything
  </h1>
  <p class="max-w-2xl text-base text-neutral-600 dark:text-neutral-400 mb-10">
    An open-source AI chat interface that gives you free access to the best models.
  </p>

  <div class="w-full max-w-4xl mb-12">
    <div class="relative w-full">
      <textarea
        bind:value={$promptText}
        on:input={(e) => updatePrompt((e.target as HTMLTextAreaElement).value)}
        on:keydown={(e) => {
          if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSend();
          }
        }}
        rows="5"
        placeholder="Ask anything or describe what you want to create..."
        class="w-full pl-6 pr-20 py-5 bg-white dark:bg-black rounded-2xl text-lg outline-none resize-none align-bottom placeholder-neutral-500 dark:placeholder-neutral-600 border border-neutral-200 dark:border-neutral-800 focus:ring-2 focus:ring-black dark:focus:ring-white transition-shadow duration-300"
        aria-label="Enter your prompt"
      ></textarea>
      <button
        on:click={handleSend}
        disabled={!($promptText || '').trim() || isSubmitting}
        class="absolute top-5 right-5 w-10 h-10 flex items-center justify-center bg-black dark:bg-white text-white dark:text-black font-bold rounded-full disabled:opacity-30 disabled:cursor-not-allowed transition-all duration-300 hover:scale-105"
        aria-label="Submit prompt"
      >
        {@html icons.arrowUp}
      </button>
    </div>

    <div class="mt-4 flex items-center gap-3">
      <div class="flex-grow"></div>
      {#if !$isLoading && !$error && $models.length > 0}
        <div class="relative" bind:this={dropdownRef}>
          <button
            on:click={() => isDropdownOpen = !isDropdownOpen}
            class="flex items-center gap-2 px-4 py-2.5 text-sm font-medium bg-white dark:bg-black border border-neutral-200 dark:border-neutral-800 rounded-lg hover:border-black dark:hover:border-white transition-colors"
          >
            <span>{$models.find(m => m.id === $selectedModelId)?.name || 'Select Model'}</span>
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
                      saveSelectedModel(model.id);
                      isDropdownOpen = false;
                    }}
                    class="w-full px-4 py-2 text-left hover:bg-neutral-100 dark:hover:bg-neutral-900 {model.id === $selectedModelId ? 'bg-neutral-100 dark:bg-neutral-900' : ''}"
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
    {#if $isLoading}
      <div
        class="w-full grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5 animate-pulse"
      >
        {#each { length: 8 } as _}
          <div
            class="p-5 border border-neutral-200 dark:border-neutral-800 rounded-xl bg-white dark:bg-black"
          >
            <div
              class="w-8 h-8 mb-4 bg-neutral-100 dark:bg-neutral-900 rounded-md"
            ></div>
            <div
              class="w-3/4 h-4 mb-2 bg-neutral-100 dark:bg-neutral-900 rounded-md"
            ></div>
            <div
              class="w-full h-3 bg-neutral-100 dark:bg-neutral-900 rounded-md"
            ></div>
            <div
              class="w-1/2 h-3 mt-1.5 bg-neutral-100 dark:bg-neutral-900 rounded-md"
            ></div>
          </div>
        {/each}
      </div>
    {:else if $error}
      <div
        class="w-full text-center p-10 bg-neutral-50 dark:bg-neutral-900/50 border border-neutral-200 dark:border-neutral-800 rounded-lg"
      >
        <h3 class="font-semibold text-black dark:text-white">
          Failed to Load Models
        </h3>
        <p
          class="text-sm text-neutral-600 dark:text-neutral-400 mt-2 max-w-md mx-auto"
        >
          There was a problem fetching data from the API. This can happen if the
          service is down or there is a network issue.
        </p>
        <p class="text-xs text-neutral-500 dark:text-neutral-600 mt-4">
          <strong>Error:</strong>
          {$error}
        </p>
      </div>
    {:else}
      <div
        class="w-full grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5"
      >
        {#each $models as model}
          <button
            on:click={() => selectedModelId.set(model.id)}
            class="relative text-left p-5 border rounded-xl transition-all duration-300 ease-in-out transform hover:-translate-y-1
                {$selectedModelId === model.id
              ? 'bg-black dark:bg-white text-white dark:text-black border-black dark:border-white'
              : 'bg-white dark:bg-black border-neutral-200 dark:border-neutral-800 hover:border-black dark:hover:border-white'}"
            aria-label={`Select ${model.name}`}
          >
            <div class="relative">
              <div class="w-8 h-8 mb-4">
                {@html icons.model}
              </div>
              <h3 class="font-semibold text-base mb-1.5">{model.name}</h3>
              <p
                class="text-sm leading-relaxed line-clamp-2
                    {$selectedModelId === model.id
                  ? 'text-neutral-400 dark:text-neutral-500'
                  : 'text-neutral-600 dark:text-neutral-400'}"
              >
                {model.description}
              </p>
            </div>
          </button>
        {/each}
      </div>
    {/if}
  </div>
</div>
