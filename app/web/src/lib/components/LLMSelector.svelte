<script lang="ts">
    import { llmStore } from '$lib/stores/llm';
    import { onMount } from 'svelte';

    let searchQuery = '';
    let isOpen = false;

    $: filteredModels = $llmStore.freeModels.filter(model => 
        model.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        model.description.toLowerCase().includes(searchQuery.toLowerCase())
    );

    onMount(() => {
        if ($llmStore.models.length === 0) {
            llmStore.loadModels();
        }
    });

    function handleModelSelect(model: typeof $llmStore.freeModels[0]) {
        llmStore.selectModel(model);
        isOpen = false;
    }
</script>

<div class="flex flex-col sm:flex-row gap-4 w-full max-w-4xl mb-4">
    <div class="flex-1">
        <div class="relative">
            <button
                class="flex items-center gap-2 px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                on:click={() => isOpen = !isOpen}
            >
                <span class="truncate">
                    {$llmStore.selectedModel?.name || 'Select Model'}
                </span>
                <svg
                    class="w-5 h-5 text-gray-400"
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 20 20"
                    fill="currentColor"
                >
                    <path
                        fill-rule="evenodd"
                        d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
                        clip-rule="evenodd"
                    />
                </svg>
            </button>

            {#if isOpen}
                <div class="absolute z-10 w-full mt-2 bg-white rounded-md shadow-lg">
                    <div class="p-2">
                        <input
                            type="text"
                            bind:value={searchQuery}
                            placeholder="Search models..."
                            class="w-full px-3 py-2 text-sm bg-neutral-100 dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md focus:outline-none focus:ring-2 focus:ring-black dark:focus:ring-white"
                        />
                    </div>
                    
                    <div class="max-h-60 overflow-y-auto">
                        {#if $llmStore.isLoading}
                            <div class="px-4 py-2 text-sm text-gray-500">Loading models...</div>
                        {:else if $llmStore.error}
                            <div class="px-4 py-2 text-sm text-red-500">{$llmStore.error}</div>
                        {:else if filteredModels.length === 0}
                            <div class="px-4 py-2 text-sm text-gray-500">No models found</div>
                        {:else}
                            {#each filteredModels as model}
                                <button
                                    class="w-full px-4 py-2 text-left hover:bg-gray-100 focus:outline-none focus:bg-gray-100"
                                    on:click={() => handleModelSelect(model)}
                                >
                                    <div class="font-medium text-gray-900">{model.name}</div>
                                    <div class="text-sm text-gray-500">{model.description}</div>
                                    <div class="text-xs text-gray-400">
                                        Context: {model.context_length.toLocaleString()} tokens
                                    </div>
                                </button>
                            {/each}
                        {/if}
                    </div>
                </div>
            {/if}
        </div>
    </div>
</div>

{#if $llmStore.error}
    <div class="w-full max-w-4xl mb-4 p-4 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-xl">
        {$llmStore.error}
    </div>
{/if} 