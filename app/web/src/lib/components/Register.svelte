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

<div class="min-h-[calc(100vh-7rem)] flex items-center justify-center px-4">
  <div class="w-full max-w-md glass rounded-2xl p-6 md:p-8 elev-1">
    <h1 class="headline text-center mb-6">Create an account</h1>
    <div class="space-y-3">
      <button
        type="button"
        on:click={() => handleSocialLogin("google")}
        disabled={$auth.isLoading}
        class="w-full flex items-center justify-center gap-3 py-3 px-4 rounded-full border border-neutral-200/70 dark:border-neutral-800/70 hover:bg-black/5 dark:hover:bg-white/5 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        aria-label="Sign up with Google"
      >
        {@html icons.google}
        <span>Sign up with Google</span>
      </button>
      <button
        type="button"
        on:click={() => handleSocialLogin("github")}
        disabled={$auth.isLoading}
        class="w-full flex items-center justify-center gap-3 py-3 px-4 rounded-full border border-neutral-200/70 dark:border-neutral-800/70 hover:bg-black/5 dark:hover:bg-white/5 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        aria-label="Sign up with GitHub"
      >
        {@html icons.github}
        <span>Sign up with GitHub</span>
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
            class="input rounded-xl bg-white/70 dark:bg-black/40 border-neutral-200/70 dark:border-neutral-800/70 placeholder-neutral-500 dark:placeholder-neutral-500"
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
            class="input rounded-xl bg-white/70 dark:bg-black/40 border-neutral-200/70 dark:border-neutral-800/70 placeholder-neutral-500 dark:placeholder-neutral-500"
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
            class="input rounded-xl bg-white/70 dark:bg-black/40 border-neutral-200/70 dark:border-neutral-800/70 placeholder-neutral-500 dark:placeholder-neutral-500"
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
          class="w-full flex justify-center py-3 px-4 rounded-full text-sm font-semibold text-white bg-black dark:bg-white dark:text-black hover:opacity-90 focus:outline-none focus:ring-2 focus:ring-black dark:focus:ring-white disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
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
