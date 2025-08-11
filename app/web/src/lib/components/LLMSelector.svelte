<script lang="ts">
    import { llmStore } from '$lib/stores/llm';
    import { onMount } from 'svelte';

    export let showTopModels: boolean = true;
    export let minimal: boolean = false;

    let searchQuery = '';
    let isOpen = false;

    // Top models to show prominently - GPT OSS 20B first as default
    const TOP_MODELS = [
        'gpt-oss-20b',
        'mistralai/mistral-7b-instruct',
        'google/gemma-7b-it',
        'meta-llama/Llama-2-7b-chat-hf',
        'deepseek/deepseek-chat:free'
    ];

    $: topModels = TOP_MODELS.map(id => 
        $llmStore.freeModels.find(model => model.id === id)
    ).filter(Boolean);

    $: filteredModels = searchQuery === '' 
        ? $llmStore.freeModels
        : $llmStore.freeModels.filter(model => 
            model.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
            model.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
            model.id.toLowerCase().includes(searchQuery.toLowerCase())
        );

    onMount(() => {
        if ($llmStore.models.length === 0) {
            llmStore.loadModels();
        }
        
        // Ensure default model is selected if none is selected
        if (!$llmStore.selectedModel && $llmStore.freeModels.length > 0) {
            const defaultModel = $llmStore.freeModels.find(m => m.id === 'gpt-oss-20b') || $llmStore.freeModels[0];
            if (defaultModel) {
                llmStore.selectModel(defaultModel);
            }
        }
    });

    function handleModelSelect(model: typeof $llmStore.freeModels[0]) {
        llmStore.selectModel(model);
        isOpen = false;
    }

    function handleTopModelSelect(model: typeof $llmStore.freeModels[0]) {
        llmStore.selectModel(model);
        isOpen = false;
    }

    function isModelSelected(model: typeof $llmStore.freeModels[0]): boolean {
        return $llmStore.selectedModel?.id === model.id;
    }
</script>

<div class={minimal ? 'inline-block' : 'flex flex-col gap-4 w-full max-w-4xl mb-4'}>
    
    {#if showTopModels && topModels.length > 0}
        <div class="mb-2">
            <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">Top Models</h3>
            <div class="flex flex-wrap gap-2">
                {#each topModels as model}
                    <button
                        class="px-3 py-2 text-sm font-medium rounded-full transition-all duration-200 {isModelSelected(model)
                            ? 'bg-black dark:bg-white text-white dark:text-black shadow-lg' 
                            : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'}"
                        on:click={() => handleTopModelSelect(model)}
                    >
                        {model.name}
                    </button>
                {/each}
            </div>
        </div>
    {/if}

    
    <div class="flex-1">
        <div class="relative">
            <button
                class="flex items-center gap-2 px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 glass rounded-full hover:opacity-95 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-black dark:focus:ring-white"
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
                <div class="absolute z-10 mt-2 backdrop-blur-3xl rounded-2xl shadow-xl elev-1 right-0" style="width: {minimal ? 'min(90vw, 48rem)' : '100%'}">
                    <div class="p-2">
                        <input
                            type="text"
                            bind:value={searchQuery}
                            placeholder="Search models by name, description, or ID..."
                            class="w-full px-3 py-2 text-sm bg-white/60 dark:bg-black/50 border border-neutral-200/60 dark:border-neutral-800/60 rounded-full focus:outline-none focus:ring-2 focus:ring-black dark:focus:ring-white"
                        />
                    </div>
                    
                    <div class="max-h-[70vh] overflow-y-auto">
                        {#if $llmStore.isLoading}
                            <div class="px-4 py-2 text-sm text-gray-500">Loading models...</div>
                        {:else if filteredModels.length === 0}
                            <div class="px-4 py-2 text-sm text-gray-500">No models found</div>
                        {:else}
                            {#each filteredModels as model}
                                <button
                                    class="w-full px-4 py-2 text-left rounded-xl hover:bg-black/5 dark:hover:bg-white/5 focus:outline-none {isModelSelected(model) ? 'bg-black/10 dark:bg-white/10 border-l-4 border-l-black dark:border-l-white' : ''}"
                                    on:click={() => handleModelSelect(model)}
                                >
                                    <div class="font-medium text-gray-900 dark:text-gray-100">{model.name}</div>
                                    <div class="text-sm text-gray-500 dark:text-gray-400 line-clamp-2">{model.description}</div>
                                    <div class="text-xs text-gray-400 dark:text-gray-500 mt-1">
                                        <span class="bg-gray-100 dark:bg-natural-800 px-2 py-1 rounded text-xs">
                                            {model.id}
                                        </span>
                                        <span class="ml-2">Context: {model.context_length.toLocaleString()} tokens</span>
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