<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import { auth } from '$lib/stores/auth';
  import { theme } from '$lib/stores/theme';
  import type { Theme } from '$lib/stores/theme';
  import { icons } from '$lib/icons';
  import ErrorMessage from './ErrorMessage.svelte';
  import { ApiError, ErrorCodes } from '$lib/api/errors';

  const isValidTheme = (value: string): value is Theme => {
    return ['light', 'dark', 'system'].includes(value);
  };

  interface ProfileResponse {
    message?: string;
    user?: {
      id: string;
      name: string;
      avatarUrl: string;
      preferences: {
        theme: Theme;
        language: string;
        timezone: string;
        notifications: boolean;
      };
    };
  }

  interface Profile {
    name: string;
    avatarUrl: string;
    preferences: {
      theme: string;
      language: string;
      timezone: string;
      notifications: boolean;
    };
  }

  let name = '';
  let avatarUrl = '';
  let isSubmitting = false;
  let error: ApiError | null = null;
  let isSuccess = false;
  let successMessage = '';
  let selectedTheme: Theme = 'system';
  let avatarFile: File | null = null;
  let avatarPreview = '';
  let nameError = '';
  let avatarError = '';

  const validateName = (value: string) => {
    if (!value.trim()) {
      return 'Name is required';
    }
    if (value.length < 2) {
      return 'Name must be at least 2 characters long';
    }
    if (value.length > 50) {
      return 'Name must be less than 50 characters';
    }
    return '';
  };

  const validateAvatar = (file: File | null) => {
    if (!file) return '';
    
    // Check file size (5MB max)
    if (file.size > 5 * 1024 * 1024) {
      return 'File size must be less than 5MB';
    }

    // Check file type
    const validTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
    if (!validTypes.includes(file.type)) {
      return 'File must be an image (JPEG, PNG, GIF, or WebP)';
    }

    return '';
  };

  onMount(async () => {
    try {
      const response = await api.getProfile();
      name = response.name || '';
      avatarUrl = response.avatarUrl || '';
      const storedTheme = localStorage.getItem('theme');
      const themeFromResponse = response.preferences?.theme;
      selectedTheme = (themeFromResponse && isValidTheme(themeFromResponse)) 
        ? themeFromResponse 
        : (storedTheme && isValidTheme(storedTheme) ? storedTheme : 'system');
      if (avatarUrl) {
        avatarPreview = avatarUrl;
      }
    } catch (err) {
      error = new ApiError('Failed to load profile', 500, ErrorCodes.UNKNOWN_ERROR);
    }
  });

  function handleThemeChange(newTheme: Theme) {
    selectedTheme = newTheme;
    theme.set(newTheme);
  }

  function handleAvatarChange(event: Event) {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files[0]) {
      const file = input.files[0];
      avatarError = validateAvatar(file);
      if (!avatarError) {
        avatarFile = file;
        const reader = new FileReader();
        reader.onload = (e) => {
          avatarPreview = e.target?.result as string;
        };
        reader.readAsDataURL(avatarFile);
      } else {
        input.value = ''; // Clear the input
      }
    }
  }

  function removeAvatar() {
    avatarFile = null;
    avatarUrl = '';
    avatarPreview = '';
    avatarError = '';
  }

  async function handleSubmit() {
    isSubmitting = true;
    error = null;
    isSuccess = false;
    successMessage = '';
    nameError = '';
    avatarError = '';

    // Validate name
    nameError = validateName(name);
    if (nameError) {
      isSubmitting = false;
      return;
    }

    // Validate avatar if there's a new one
    if (avatarFile) {
      avatarError = validateAvatar(avatarFile);
      if (avatarError) {
        isSubmitting = false;
        return;
      }
    }

    try {
      // First upload avatar if there's a new one
      if (avatarFile) {
        const formData = new FormData();
        formData.append('avatar', avatarFile);
        const uploadResult = await api.uploadAvatar(formData);
        avatarUrl = uploadResult.url;
      }

      // Then update profile
      const response = await api.updateProfile({ 
        name, 
        avatarUrl: avatarUrl,
        preferences: {
          theme: selectedTheme
        }
      }) as ProfileResponse;

      isSuccess = true;
      successMessage = response.message || 'Profile updated successfully';

      // Update local theme if changed
      if (response.user?.preferences?.theme) {
        theme.set(response.user.preferences.theme);
      }
    } catch (err: unknown) {
      if (err instanceof ApiError) {
        error = err;
      } else if (err instanceof Error) {
        error = new ApiError(err.message, 500, ErrorCodes.UNKNOWN_ERROR);
      } else {
        error = new ApiError('An unexpected error occurred', 500, ErrorCodes.UNKNOWN_ERROR);
      }
    } finally {
      isSubmitting = false;
    }
  }
</script>

<div class="max-w-2xl mx-auto p-6">
  <div class="bg-white dark:bg-zinc-900 rounded-lg shadow-sm border border-neutral-200 dark:border-neutral-800">
    <div class="p-6">
      <h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-6">Profile Settings</h2>
      
      <form on:submit|preventDefault={handleSubmit} class="space-y-6">
        <!-- Avatar Section -->
        <div class="flex items-center space-x-6">
          <div class="shrink-0">
            {#if avatarPreview}
              <img
                src={avatarPreview}
                alt="Profile"
                class="h-24 w-24 rounded-full object-cover"
              />
            {:else}
              <div class="h-24 w-24 rounded-full bg-gray-200 dark:bg-natural-900 flex items-center justify-center">
                <span class="text-2xl font-medium text-gray-500 dark:text-gray-400">
                  {name?.[0] || '?'}
                </span>
              </div>
            {/if}
          </div>
          <div class="flex flex-col space-y-2">
            <label for="avatar-upload" class="block">
              <span class="sr-only">Choose profile photo</span>
              <input
                id="avatar-upload"
                type="file"
                accept="image/jpeg,image/png,image/gif,image/webp"
                on:change={handleAvatarChange}
                class="block w-full text-sm text-gray-500 dark:text-gray-400
                  file:mr-4 file:py-2 file:px-4
                  file:rounded-md file:border-0
                  file:text-sm file:font-medium
                  file:bg-black dark:file:bg-white
                  file:text-white dark:file:text-black
                  hover:file:opacity-90"
              />
            </label>
            {#if avatarError}
              <div class="text-red-500 text-sm">{avatarError}</div>
            {/if}
            {#if avatarPreview}
              <button
                type="button"
                on:click={removeAvatar}
                class="text-sm text-red-600 dark:text-red-400 hover:underline"
              >
                Remove avatar
              </button>
            {/if}
          </div>
        </div>

        <!-- Name Field -->
        <div>
          <label for="name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Name
          </label>
          <input
            type="text"
            id="name"
            bind:value={name}
            class="w-full px-3 py-2 border border-gray-300 dark:border-neutral-700 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-black dark:focus:ring-white focus:border-black dark:focus:border-white sm:text-sm bg-white dark:bg-neutral-800"
            placeholder="Your name"
          />
          {#if nameError}
            <div class="text-red-500 text-sm mt-1">{nameError}</div>
          {/if}
        </div>

        <!-- Theme Preferences -->
        <div>
          <fieldset>
            <legend class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Theme Preference
            </legend>
            <div class="grid grid-cols-3 gap-3">
              <label class="relative flex items-center justify-center gap-2 px-4 py-2 border rounded-md cursor-pointer {$theme === 'light' ? 'border-black dark:border-white bg-black dark:bg-white text-white dark:text-black' : 'border-gray-300 dark:border-neutral-700 hover:border-black dark:hover:border-white'}">
                <input
                  type="radio"
                  name="theme"
                  value="light"
                  checked={selectedTheme === 'light'}
                  on:change={() => handleThemeChange('light')}
                  class="sr-only"
                />
                {@html icons.sun}
                <span>Light</span>
              </label>
              <label class="relative flex items-center justify-center gap-2 px-4 py-2 border rounded-md cursor-pointer {$theme === 'dark' ? 'border-black dark:border-white bg-black dark:bg-white text-white dark:text-black' : 'border-gray-300 dark:border-neutral-700 hover:border-black dark:hover:border-white'}">
                <input
                  type="radio"
                  name="theme"
                  value="dark"
                  checked={selectedTheme === 'dark'}
                  on:change={() => handleThemeChange('dark')}
                  class="sr-only"
                />
                {@html icons.moon}
                <span>Dark</span>
              </label>
              <label class="relative flex items-center justify-center gap-2 px-4 py-2 border rounded-md cursor-pointer {$theme === 'system' ? 'border-black dark:border-white bg-black dark:bg-white text-white dark:text-black' : 'border-gray-300 dark:border-neutral-700 hover:border-black dark:hover:border-white'}">
                <input
                  type="radio"
                  name="theme"
                  value="system"
                  checked={selectedTheme === 'system'}
                  on:change={() => handleThemeChange('system')}
                  class="sr-only"
                />
                {@html icons.system}
                <span>System</span>
              </label>
            </div>
          </fieldset>
        </div>

        {#if error}
          <ErrorMessage {error} onRetry={handleSubmit} />
        {/if}

        {#if isSuccess}
          <div class="text-green-500 text-sm">{successMessage}</div>
        {/if}

        <div class="flex justify-end">
          <button
            type="submit"
            disabled={isSubmitting}
            class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-black dark:text-black dark:bg-white hover:opacity-90 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-black dark:focus:ring-white disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
          >
            {#if isSubmitting}
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
              Saving...
            {:else}
              Save Changes
            {/if}
          </button>
        </div>
      </form>
    </div>
  </div>
</div> 