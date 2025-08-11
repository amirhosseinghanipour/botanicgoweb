<script lang="ts">
  import { isLoggedIn } from "$lib/stores";
  import { api } from "$lib/api/client";
  import { goto } from '$app/navigation';
  import { auth } from '$lib/stores/auth';
  import ErrorMessage from './ErrorMessage.svelte';
  import { ApiError, ErrorCodes } from '$lib/api/errors';
  import { icons } from "$lib/icons.js";
  import { onMount } from "svelte";

  let email = "";
  let password = "";
  let rememberMe = false;
  let loading = false;
  let error: ApiError | null = null;
  let formElement: HTMLFormElement;
  let emailInput: HTMLInputElement;
  
  function validateEmail(value: string): boolean {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
  }
  function validatePassword(value: string): boolean {
    return value.length >= 6;
  }
  $: isFormValid = validateEmail(email) && validatePassword(password);

  // Focus email input on mount for better UX
  onMount(() => {
    emailInput?.focus();
  });

  async function handleSubmit(event?: SubmitEvent) {
    if (event) {
      event.preventDefault();
    }
    if (!isFormValid || loading) {
      return;
    }
    loading = true;
    error = null;

    try {
      const response = await api.login({ email, password, rememberMe });
      auth.setUser(response.user);
      goto('/chat');
    } catch (err: unknown) {
      if (err instanceof ApiError) {
        error = err;
      } else if (err instanceof Error) {
        error = new ApiError(err.message, 500, ErrorCodes.UNKNOWN_ERROR);
      } else {
        error = new ApiError('An unexpected error occurred', 500, ErrorCodes.UNKNOWN_ERROR);
      }
    } finally {
      loading = false;
    }
  }

  const handleGoogleLogin = () => {
    window.location.href = 'http://localhost:8000/api/auth/google';
  };

  const handleGithubLogin = () => {
    window.location.href = 'http://localhost:8000/api/auth/github';
  };
</script>

<div class="min-h-[calc(100vh-7rem)] flex items-center justify-center px-4">
  <div class="w-full max-w-md glass rounded-2xl p-6 md:p-8 elev-1">
    <h1 class="headline text-center mb-6">Sign in</h1>
    <div class="space-y-3">
      <button
        type="button"
        on:click={handleGoogleLogin}
        disabled={loading}
        class="w-full flex items-center justify-center gap-3 py-3 px-4 rounded-full border border-neutral-200/70 dark:border-neutral-800/70 hover:bg-black/5 dark:hover:bg-white/5 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        aria-label="Sign in with Google"
      >
        {@html icons.google}
        <span>Sign in with Google</span>
      </button>
      <button
        type="button"
        on:click={handleGithubLogin}
        disabled={loading}
        class="w-full flex items-center justify-center gap-3 py-3 px-4 rounded-full border border-neutral-200/70 dark:border-neutral-800/70 hover:bg-black/5 dark:hover:bg-white/5 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        aria-label="Sign in with GitHub"
      >
        {@html icons.github}
        <span>Sign in with GitHub</span>
      </button>
    </div>
    <div class="my-6 separator"></div>
    <form
      bind:this={formElement}
      on:submit={handleSubmit}
      class="space-y-4"
      novalidate
    >
      <div>
        <label
          for="email-login"
          class="block text-sm font-medium text-left text-gray-700 dark:text-gray-300"
          >Email address</label
        >
        <div class="mt-1">
          <input
            bind:this={emailInput}
            id="email-login"
            name="email"
            type="email"
            bind:value={email}
            autocomplete="email"
            required
            class="input rounded-xl bg-white/70 dark:bg-black/40 border-neutral-200/70 dark:border-neutral-800/70 placeholder-neutral-500 dark:placeholder-neutral-500"
            aria-describedby={error ? "login-error" : undefined}
            disabled={loading}
            placeholder="Enter your email"
          />
        </div>
      </div>

      <div>
        <label
          for="password"
          class="block text-sm font-medium text-left text-gray-700 dark:text-gray-300"
          >Password</label
        >
        <div class="mt-1">
          <input
            id="password"
            name="password"
            type="password"
            bind:value={password}
            autocomplete="current-password"
            required
            class="input rounded-xl bg-white/70 dark:bg-black/40 border-neutral-200/70 dark:border-neutral-800/70 placeholder-neutral-500 dark:placeholder-neutral-500"
            aria-describedby={error ? "login-error" : undefined}
            disabled={loading}
            placeholder="Enter your password"
          />
        </div>
      </div>

      <div class="flex items-center justify-between">
        <div class="flex items-center">
          <input
            id="remember-me"
            name="remember-me"
            type="checkbox"
            bind:checked={rememberMe}
            class="h-4 w-4 text-black dark:text-white focus:ring-black dark:focus:ring-white border-gray-300 dark:border-gray-700 rounded bg-gray-100 dark:bg-neutral-900/50"
          />
          <label for="remember-me" class="ml-2 block text-sm text-gray-700 dark:text-gray-300">
            Remember me
          </label>
        </div>
      </div>

      <div>
        <button
          type="submit"
          disabled={loading || !isFormValid}
          class="w-full flex justify-center py-3 px-4 rounded-full text-sm font-semibold text-white bg-black dark:bg-white dark:text-black hover:opacity-90 focus:outline-none focus:ring-2 focus:ring-black dark:focus:ring-white disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {#if loading}
            <svg
              class="animate-spin -ml-1 mr-3 h-5 w-5 text-white dark:text-black"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
          {/if}
          Sign in
        </button>
      </div>
    </form>

    <ErrorMessage {error} onRetry={handleSubmit} />

    <div class="mt-6 text-center">
      <p class="text-sm">
        Do not have an account?
        <a
          href="/register"
          class="font-semibold text-black dark:text-white hover:underline focus:outline-none focus:ring-2 focus:ring-black dark:focus:ring-white"
        >
          Register
        </a>
      </p>
    </div>
  </div>
</div>