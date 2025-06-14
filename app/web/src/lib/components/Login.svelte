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

  // Focus email input on mount for better UX
  onMount(() => {
    emailInput?.focus();
  });

  async function handleSubmit(event?: SubmitEvent) {
    if (event) {
      event.preventDefault();
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

<div class="grid md:grid-cols-2 gap-8 items-center">
  <div
    class="relative h-full w-full rounded-lg overflow-hidden hidden md:block"
    role="img"
    aria-label="Abstract decorative background image"
  >
    <img
      src="https://images.unsplash.com/photo-1556139930-c23fa4a4f934?q=80&w=1470&auto=format&fit=crop&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D"
      alt="Abstract background"
      class="h-full w-full object-cover"
      loading="lazy"
      decoding="async"
    />
    <div
      class="absolute inset-0 bg-black/30 backdrop-blur-sm flex flex-col justify-center p-8"
    >
      <h2 class="text-4xl font-bold text-white mb-4">Welcome Back</h2>
      <p class="text-white/80">
        Sign in to continue your journey with AI-powered development.
      </p>
    </div>
  </div>
  <div class="flex flex-col justify-center">
    <h1
      class="text-3xl md:text-4xl font-bold tracking-tighter mb-2 text-center"
    >
      Sign In
    </h1>
    <p class="text-gray-500 dark:text-gray-400 mb-6 text-center">
      Welcome back! Please enter your details.
    </p>
    <div class="space-y-4">
      <button
        type="button"
        on:click={handleGoogleLogin}
        disabled={loading}
        class="w-full flex items-center justify-center gap-3 py-3 px-4 border border-gray-300 dark:border-gray-700 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        aria-label="Sign in with Google"
      >
        {@html icons.google}
        <span>Sign in with Google</span>
      </button>
      <button
        type="button"
        on:click={handleGithubLogin}
        disabled={loading}
        class="w-full flex items-center justify-center gap-3 py-3 px-4 border border-gray-300 dark:border-gray-700 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        aria-label="Sign in with GitHub"
      >
        {@html icons.github}
        <span>Sign in with GitHub</span>
      </button>
    </div>
    <div
      class="my-6 flex items-center before:flex-1 before:border-t before:border-gray-300 dark:before:border-gray-700 after:flex-1 after:border-t after:border-gray-300 dark:after:border-gray-700"
      role="separator"
    >
      <p class="mx-4 text-center text-sm text-gray-500 dark:text-gray-400">
        OR
      </p>
    </div>
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
            class="appearance-none block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md shadow-sm placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-black dark:focus:ring-white focus:border-black dark:focus:border-white sm:text-sm bg-gray-100 dark:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed"
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
            class="appearance-none block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md shadow-sm placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-black dark:focus:ring-white focus:border-black dark:focus:border-white sm:text-sm bg-gray-100 dark:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed"
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
          disabled={loading}
          class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-black dark:bg-white dark:text-black hover:bg-gray-800 dark:hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-black dark:focus:ring-white disabled:opacity-50 disabled:cursor-not-allowed"
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
          class="font-medium text-black dark:text-white hover:underline focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-black dark:focus:ring-white"
        >
          Register
        </a>
      </p>
    </div>
  </div>
</div>