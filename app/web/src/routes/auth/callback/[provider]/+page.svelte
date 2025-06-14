<script lang="ts">
    import type { PageData } from './$types';
    import OAuthCallback from '$lib/components/auth/OAuthCallback.svelte';
    import { page } from '$app/stores';
</script>

<div class="flex flex-col items-center justify-center min-h-screen p-4">
    <div class="text-center">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-black dark:border-white mx-auto mb-4"></div>
        <h1 class="text-2xl font-bold mb-2">Completing Authentication</h1>
        <p class="text-gray-600 dark:text-gray-400">
            Please wait while we complete your {#if $page.params.provider === 'google'}Google{:else}GitHub{/if} sign-in...
        </p>
    </div>
</div>

<OAuthCallback provider={data.provider} /> 