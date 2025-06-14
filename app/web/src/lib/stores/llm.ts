import { writable, derived, get } from 'svelte/store';
import { browser } from '$app/environment';
import type { Writable } from 'svelte/store';

export interface Model {
  id: string;
  name: string;
  description: string;
  context_length: number;
  pricing: {
    prompt: string;
    completion: string;
  };
  architecture: {
    modality: string;
    tokenizer: string;
    instruct_type: string;
  };
}

interface LLMStoreState {
  models: Model[];
  freeModels: Model[];
  nonFreeModels: Model[];
  selectedModel: Model | null;
  isLoading: boolean;
  error: string | null;
  hasMore: boolean;
  currentPage: number;
  pageSize: number;
}

interface LLMStore extends Writable<LLMStoreState> {
  loadModels: () => Promise<void>;
  loadMoreModels: () => Promise<void>;
  selectModel: (model: Model) => void;
}

const CACHE_KEY = 'cachedModels';
const CACHE_TIMESTAMP_KEY = 'modelsCacheTimestamp';
const CACHE_DURATION_MS = 24 * 60 * 60 * 1000; // 24 hours

function createLLMStore(): LLMStore {
  const initialState: LLMStoreState = {
    models: [],
    freeModels: [],
    nonFreeModels: [],
    selectedModel: null,
    isLoading: true,
    error: null,
    hasMore: false,
    currentPage: 1,
    pageSize: 10
  };

  const store = writable<LLMStoreState>(initialState);

  const loadFromCache = () => {
    if (!browser) return false;
    
    const cachedData = localStorage.getItem(CACHE_KEY);
    if (cachedData) {
      try {
        const parsedData = JSON.parse(cachedData);
        store.update(state => ({
          ...state,
          models: parsedData.all || [],
          freeModels: parsedData.free || [],
          nonFreeModels: parsedData.nonFree || [],
          isLoading: false
        }));
        return true;
      } catch (e) {
        console.warn('Failed to parse cache:', e);
        return false;
      }
    }
    return false;
  };

  const loadModels = async () => {
    try {
      const response = await fetch('/api/models?page=1&pageSize=10');
      if (!response.ok) {
        throw new Error('Failed to fetch models');
      }
      const data = await response.json();
      if (data.success) {
        store.update(state => ({
          ...state,
          freeModels: data.data.free || [],
          nonFreeModels: data.data.nonFree || [],
          hasMore: data.data.hasMore,
          currentPage: data.data.page,
          pageSize: data.data.pageSize,
          isLoading: false
        }));
      } else {
        throw new Error(data.error || 'Failed to fetch models');
      }
    } catch (error) {
      store.update(state => ({
        ...state,
        error: error instanceof Error ? error.message : 'Failed to fetch models',
        isLoading: false
      }));
    }
  };

  const loadMoreModels = async () => {
    const state = get(store);
    if (state.isLoading || !state.hasMore) return;

    store.update(state => ({ ...state, isLoading: true }));
    try {
      const nextPage = state.currentPage + 1;
      const response = await fetch(`/api/models?page=${nextPage}&pageSize=${state.pageSize}`);
      if (!response.ok) {
        throw new Error('Failed to fetch more models');
      }
      const data = await response.json();
      if (data.success) {
        store.update(state => ({
          ...state,
          nonFreeModels: [...state.nonFreeModels, ...(data.data.nonFree || [])],
          hasMore: data.data.hasMore,
          currentPage: data.data.page,
          isLoading: false
        }));
      } else {
        throw new Error(data.error || 'Failed to fetch more models');
      }
    } catch (error) {
      store.update(state => ({
        ...state,
        error: error instanceof Error ? error.message : 'Failed to fetch more models',
        isLoading: false
      }));
    }
  };

  const selectModel = (model: Model) => {
    store.update(state => ({ ...state, selectedModel: model }));
    if (browser && model) {
      localStorage.setItem('selectedModelId', model.id);
    }
  };

  return {
    ...store,
    loadModels,
    loadMoreModels,
    selectModel
  };
}

export const llmStore = createLLMStore(); 