<script lang="ts">
  import { ApiError, NetworkError, ErrorCodes, type ErrorCode } from '$lib/api/errors';

  export let error: Error | null = null;
  export let onRetry: (() => void) | null = null;

  $: errorMessage = getErrorMessage(error);
  $: showRetry = error instanceof NetworkError || (error instanceof ApiError && error.code === ErrorCodes.SERVER_ERROR);

  function getErrorMessage(err: Error | null): string {
    if (!err) return '';

    if (err instanceof NetworkError) {
      return 'Network error occurred. Please check your internet connection.';
    }

    if (err instanceof ApiError) {
      const code = err.code as ErrorCode;
      switch (code) {
        case ErrorCodes.SESSION_EXPIRED:
          return 'Your session has expired. Please log in again.';
        case ErrorCodes.NO_TOKEN:
          return 'You are not logged in. Please log in to continue.';
        case ErrorCodes.REFRESH_FAILED:
          return 'Failed to refresh your session. Please log in again.';
        case ErrorCodes.VALIDATION_ERROR:
          return err.details?.message || 'Please check your input and try again.';
        case ErrorCodes.RATE_LIMIT:
          return 'Too many requests. Please try again later.';
        case ErrorCodes.SERVER_ERROR:
          return 'Server error occurred. Please try again later.';
        default:
          return err.message || 'An unexpected error occurred.';
      }
    }

    return err.message || 'An unexpected error occurred.';
  }
</script>

{#if error}
  <div class="rounded-md bg-red-50 dark:bg-red-900/50 p-4">
    <div class="flex">
      <div class="flex-shrink-0">
        <svg
          class="h-5 w-5 text-red-400 dark:text-red-300"
          viewBox="0 0 20 20"
          fill="currentColor"
        >
          <path
            fill-rule="evenodd"
            d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
            clip-rule="evenodd"
          />
        </svg>
      </div>
      <div class="ml-3">
        <h3 class="text-sm font-medium text-red-800 dark:text-red-200">
          {errorMessage}
        </h3>
        {#if showRetry && onRetry}
          <div class="mt-2">
            <button
              type="button"
              class="inline-flex items-center px-2.5 py-1.5 border border-transparent text-xs font-medium rounded text-red-700 dark:text-red-200 bg-red-100 dark:bg-red-800 hover:bg-red-200 dark:hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
              on:click={() => onRetry?.()}
            >
              Try again
            </button>
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if} 