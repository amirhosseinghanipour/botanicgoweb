<script lang="ts">
  import { icons } from "$lib/icons.js";
  import { goto } from "$app/navigation";
  import { auth } from "$lib/stores/auth";
  import { api } from "$lib/api/client";
  import { onMount } from "svelte";
  import ErrorMessage from './ErrorMessage.svelte';
  import { ApiError, ErrorCodes } from '$lib/api/errors';

  let email = "";
  let password = "";
  let confirmPassword = "";
  let acceptTerms = false;
  let formElement: HTMLFormElement;
  let isSubmitting = false;
  let emailInput: HTMLInputElement;
  let name = "";
  let error: ApiError | null = null;
  let loading = false;

  // Focus email input on mount for better UX
  onMount(() => {
    emailInput?.focus();
  });

  async function handleSubmit(event?: SubmitEvent) {
    if (event) {
      event.preventDefault();
    }
    if (isSubmitting) return;

    if (password !== confirmPassword) {
      auth.setError("Passwords do not match");
      return;
    }

    if (!acceptTerms) {
      auth.setError("Please accept the terms and privacy policy");
      return;
    }

    try {
      isSubmitting = true;
      auth.setLoading(true);
      auth.setError("Error...s");

      const response = await api.register({ email, password });
      auth.setUser(response.user);
      await goto("/chat");
    } catch (err: unknown) {
      if (err instanceof ApiError) {
        error = err;
      } else if (err instanceof Error) {
        error = new ApiError(err.message, 500, ErrorCodes.UNKNOWN_ERROR);
      } else {
        error = new ApiError('An unexpected error occurred', 500, ErrorCodes.UNKNOWN_ERROR);
      }
      auth.setError(error?.message || 'An error occurred');
      formElement.reportValidity();
    } finally {
      isSubmitting = false;
      auth.setLoading(false);
    }
  }

  async function handleSocialLogin(provider: "google" | "github") {
    try {
      auth.setLoading(true);
      auth.setError("Error...");
      const url = provider === "google" ? await api.getGoogleAuthUrl() : await api.getGithubAuthUrl();
      window.location.href = url;
    } catch (err) {
      const error = err as Error;
      auth.setError(error.message);
    }
  }

  // Client-side form validation
  function validateEmail(value: string): boolean {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
  }

  function validatePassword(value: string): boolean {
    return value.length >= 6;
  }

  function validateConfirmPassword(pass: string, confirm: string): boolean {
    return pass === confirm && pass.length > 0;
  }

  $: isFormValid = validateEmail(email) && 
                   validatePassword(password) && 
                   validateConfirmPassword(password, confirmPassword) && 
                   acceptTerms;
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
      <h2 class="text-4xl font-bold text-white mb-4">Unlock Your Potential</h2>
      <p class="text-white/80">
        Join a community of developers and creators pushing the boundaries of
        AI.
      </p>
    </div>
  </div>
  <div class="flex flex-col justify-center">
    <h1
      class="text-3xl md:text-4xl font-bold tracking-tighter mb-2 text-center"
    >
      Create an Account
    </h1>
    <p class="text-gray-500 dark:text-gray-400 mb-6 text-center">
      Join us and start exploring the future of AI.
    </p>
    <div class="space-y-4">
      <button
        type="button"
        on:click={() => handleSocialLogin("google")}
        disabled={$auth.isLoading}
        class="w-full flex items-center justify-center gap-3 py-3 px-4 border border-gray-300 dark:border-gray-700 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        aria-label="Sign up with Google"
      >
        {@html icons.google}
        <span>Sign up with Google</span>
      </button>
      <button
        type="button"
        on:click={() => handleSocialLogin("github")}
        disabled={$auth.isLoading}
        class="w-full flex items-center justify-center gap-3 py-3 px-4 border border-gray-300 dark:border-gray-700 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        aria-label="Sign up with GitHub"
      >
        {@html icons.github}
        <span>Sign up with GitHub</span>
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
          for="email-register"
          class="block text-sm font-medium text-left text-gray-700 dark:text-gray-300"
          >Email address</label
        >
        <div class="mt-1">
          <input
            bind:this={emailInput}
            id="email-register"
            name="email"
            type="email"
            bind:value={email}
            autocomplete="email"
            required
            pattern="[^@\s]+@[^@\s]+\.[^@\s]+"
            class="appearance-none block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md shadow-sm placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-black dark:focus:ring-white focus:border-black dark:focus:border-white sm:text-sm bg-gray-100 dark:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed"
            aria-describedby={$auth.error ? "register-error" : undefined}
            disabled={$auth.isLoading}
            placeholder="you@example.com"
          />
        </div>
      </div>

      <div>
        <label
          for="password-register"
          class="block text-sm font-medium text-left text-gray-700 dark:text-gray-300"
          >Password</label
        >
        <div class="mt-1">
          <input
            id="password-register"
            name="password"
            type="password"
            bind:value={password}
            autocomplete="new-password"
            required
            minlength="6"
            class="appearance-none block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md shadow-sm placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-black dark:focus:ring-white focus:border-black dark:focus:border-white sm:text-sm bg-gray-100 dark:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed"
            aria-describedby={$auth.error ? "register-error" : undefined}
            disabled={$auth.isLoading}
            placeholder="••••••"
          />
        </div>
      </div>

      <div>
        <label
          for="confirm-password-register"
          class="block text-sm font-medium text-left text-gray-700 dark:text-gray-300"
          >Confirm Password</label
        >
        <div class="mt-1">
          <input
            id="confirm-password-register"
            name="confirmPassword"
            type="password"
            bind:value={confirmPassword}
            autocomplete="new-password"
            required
            minlength="6"
            class="appearance-none block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md shadow-sm placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-black dark:focus:ring-white focus:border-black dark:focus:border-white sm:text-sm bg-gray-100 dark:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed"
            aria-describedby={$auth.error ? "register-error" : undefined}
            disabled={$auth.isLoading}
            placeholder="••••••"
          />
        </div>
      </div>

      <div class="flex items-center">
        <input
          id="terms-register"
          name="terms"
          type="checkbox"
          bind:checked={acceptTerms}
          required
          class="h-4 w-4 text-black dark:text-white focus:ring-black dark:focus:ring-white border-gray-300 dark:border-gray-700 rounded bg-gray-100 dark:bg-neutral-900/50"
        />
        <label
          for="terms-register"
          class="ml-2 block text-sm text-gray-700 dark:text-gray-300"
        >
          I agree to the
          <a
            href="/terms"
            class="font-bold"
            target="_blank"
            rel="noopener noreferrer"
            >Terms of Service</a
          >
          and
          <a
            href="/privacy"
            class="font-bold"
            target="_blank"
            rel="noopener noreferrer"
            >Privacy Policy</a
          >
        </label>
      </div>

      <div>
        <button
          type="submit"
          class="w-full flex justify-center py-3 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-black dark:text-black dark:bg-white hover:opacity-90 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-black dark:focus:ring-white disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
          disabled={$auth.isLoading || !isFormValid}
          aria-busy={$auth.isLoading}
        >
          {#if $auth.isLoading}
            <svg
              class="animate-spin -ml-1 mr-3 h-5 w-5"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              />
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              />
            </svg>
            Creating account...
          {:else}
            Create Account
          {/if}
        </button>
      </div>
    </form>

    <ErrorMessage {error} onRetry={handleSubmit} />

    <div class="mt-6 text-center">
      <p class="text-sm">
        Already have an account?
        <a
          href="/login"
          class="font-medium text-black dark:text-white hover:underline focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-black dark:focus:ring-white"
        >
          Login
        </a>
      </p>
    </div>
  </div>
</div>
