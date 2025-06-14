<!-- ModelComparison.svelte -->
<script lang="ts">
    import { llmStore } from '$lib/stores/llm';
    import { notifications } from '$lib/stores/notifications';
    import { store as websocket } from '$lib/stores/websocket';
    import ChatMessage from './ChatMessage.svelte';
    import type { Message } from '$lib/types';

    export let prompt: string;
    export let onClose: () => void;

    let selectedModels: string[] = [];
    let responses: Record<string, Message> = {};
    let isLoading: Record<string, boolean> = {};
    let error: Record<string, string> = {};

    $: availableModels = $llmStore.models.filter(m => !selectedModels.includes(m.id));

    async function addModel(modelId: string) {
        selectedModels = [...selectedModels, modelId];
        isLoading[modelId] = true;
        error[modelId] = '';

        try {
            const response = await fetch('/api/chat', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    content: prompt,
                    model: modelId,
                }),
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            responses[modelId] = data;
        } catch (e) {
            error[modelId] = e instanceof Error ? e.message : 'Failed to get response';
            notifications.add({
                type: 'error',
                message: `Error getting response from ${$llmStore.models.find(m => m.id === modelId)?.name || modelId}`,
            });
        } finally {
            isLoading[modelId] = false;
        }
    }

    function removeModel(modelId: string) {
        selectedModels = selectedModels.filter(id => id !== modelId);
        delete responses[modelId];
        delete isLoading[modelId];
        delete error[modelId];
    }

    function copyResponse(modelId: string) {
        const response = responses[modelId]?.content;
        if (response) {
            navigator.clipboard.writeText(response);
            notifications.add({
                type: 'success',
                message: 'Response copied to clipboard',
                duration: 2000,
            });
        }
    }
</script>

<div class="comparison-modal">
    <div class="modal-header">
        <h2>Compare Model Responses</h2>
        <button class="close-button" on:click={onClose}>×</button>
    </div>

    <div class="prompt-section">
        <h3>Prompt</h3>
        <div class="prompt-content">{prompt}</div>
    </div>

    <div class="models-section">
        <div class="model-selector">
            <select
                value=""
                on:change={(e) => {
                    const modelId = e.currentTarget.value;
                    if (modelId) {
                        addModel(modelId);
                        e.currentTarget.value = '';
                    }
                }}
                disabled={availableModels.length === 0}
            >
                <option value="">Add a model to compare...</option>
                {#each availableModels as model}
                    <option value={model.id}>{model.name}</option>
                {/each}
            </select>
        </div>

        <div class="responses-grid">
            {#each selectedModels as modelId}
                {@const model = $llmStore.models.find(m => m.id === modelId)}
                <div class="response-card">
                    <div class="response-header">
                        <h3>{model?.name || modelId}</h3>
                        <div class="response-actions">
                            <button
                                class="copy-button"
                                on:click={() => copyResponse(modelId)}
                                title="Copy response"
                                disabled={!responses[modelId]}
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                                    <path d="M8 2a1 1 0 000 2h2a1 1 0 100-2H8z" />
                                    <path d="M3 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v5h-4.586l1.293-1.293a1 1 0 00-1.414-1.414l-3 3a1 1 0 000 1.414l3 3a1 1 0 001.414-1.414L10.414 13H15v3a2 2 0 01-2 2H5a2 2 0 01-2-2V5zM15 11h2a1 1 0 110 2h-2v-2z" />
                                </svg>
                            </button>
                            <button
                                class="remove-button"
                                on:click={() => removeModel(modelId)}
                                title="Remove model"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                                    <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                                </svg>
                            </button>
                        </div>
                    </div>

                    <div class="response-content">
                        {#if isLoading[modelId]}
                            <div class="loading">
                                <div class="spinner"></div>
                                <span>Generating response...</span>
                            </div>
                        {:else if error[modelId]}
                            <div class="error">
                                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                                    <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
                                </svg>
                                <span>{error[modelId]}</span>
                            </div>
                        {:else if responses[modelId]}
                            <ChatMessage message={responses[modelId]} />
                        {/if}
                    </div>
                </div>
            {/each}
        </div>
    </div>
</div>

<style>
    .comparison-modal {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background-color: rgba(0, 0, 0, 0.5);
        display: flex;
        flex-direction: column;
        z-index: 1000;
    }

    .modal-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 1rem;
        background-color: var(--bg-color, #ffffff);
        border-bottom: 1px solid var(--border-color, #e5e7eb);
    }

    .modal-header h2 {
        margin: 0;
        font-size: 1.25rem;
        font-weight: 600;
    }

    .close-button {
        background: none;
        border: none;
        font-size: 1.5rem;
        cursor: pointer;
        padding: 0.5rem;
        color: var(--text-secondary, #6b7280);
    }

    .close-button:hover {
        color: var(--text-primary, #1f2937);
    }

    .prompt-section {
        padding: 1rem;
        background-color: var(--bg-color, #ffffff);
        border-bottom: 1px solid var(--border-color, #e5e7eb);
    }

    .prompt-section h3 {
        margin: 0 0 0.5rem 0;
        font-size: 1rem;
        font-weight: 500;
        color: var(--text-secondary, #6b7280);
    }

    .prompt-content {
        padding: 1rem;
        background-color: var(--bg-secondary, #f3f4f6);
        border-radius: 0.375rem;
        white-space: pre-wrap;
    }

    .models-section {
        flex: 1;
        padding: 1rem;
        overflow-y: auto;
    }

    .model-selector {
        margin-bottom: 1rem;
    }

    .model-selector select {
        width: 100%;
        padding: 0.5rem;
        border: 1px solid var(--border-color, #e5e7eb);
        border-radius: 0.375rem;
        background-color: var(--bg-color, #ffffff);
        color: var(--text-primary, #1f2937);
    }

    .responses-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
        gap: 1rem;
    }

    .response-card {
        background-color: var(--bg-color, #ffffff);
        border: 1px solid var(--border-color, #e5e7eb);
        border-radius: 0.5rem;
        overflow: hidden;
    }

    .response-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 0.75rem;
        background-color: var(--bg-secondary, #f3f4f6);
        border-bottom: 1px solid var(--border-color, #e5e7eb);
    }

    .response-header h3 {
        margin: 0;
        font-size: 1rem;
        font-weight: 500;
    }

    .response-actions {
        display: flex;
        gap: 0.5rem;
    }

    .response-actions button {
        padding: 0.25rem;
        background: none;
        border: none;
        color: var(--text-secondary, #6b7280);
        cursor: pointer;
        border-radius: 0.25rem;
    }

    .response-actions button:hover {
        background-color: var(--hover-color, #e5e7eb);
        color: var(--text-primary, #1f2937);
    }

    .response-actions button:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    .response-actions button svg {
        width: 1.25rem;
        height: 1.25rem;
    }

    .response-content {
        padding: 1rem;
        min-height: 200px;
    }

    .loading {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        height: 100%;
        color: var(--text-secondary, #6b7280);
    }

    .spinner {
        width: 2rem;
        height: 2rem;
        border: 3px solid var(--border-color, #e5e7eb);
        border-top-color: var(--primary-color, #3b82f6);
        border-radius: 50%;
        animation: spin 1s linear infinite;
        margin-bottom: 0.5rem;
    }

    .error {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        color: var(--error-color, #ef4444);
        padding: 0.5rem;
        background-color: var(--error-bg, #fee2e2);
        border-radius: 0.375rem;
    }

    .error svg {
        width: 1.25rem;
        height: 1.25rem;
    }

    @keyframes spin {
        to {
            transform: rotate(360deg);
        }
    }
</style> 