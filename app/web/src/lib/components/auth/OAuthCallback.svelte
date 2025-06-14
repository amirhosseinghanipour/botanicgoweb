<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { auth } from '$lib/stores/auth';
    import { secureStorage } from '$lib/utils/secureStorage';
    import { browser } from '$app/environment';

    export let provider: 'google' | 'github';

    onMount(async () => {
        if (!browser) return;

        try {
            auth.setLoading(true);
            
            // Get auth token from cookie
            const authTokenCookie = document.cookie
                .split('; ')
                .find(row => row.startsWith('auth_token='));
            
            if (!authTokenCookie) {
                throw new Error('Missing auth token');
            }

            // Get user data from cookie
            const userDataCookie = document.cookie
                .split('; ')
                .find(row => row.startsWith('user_data='));
            
            if (!userDataCookie) {
                throw new Error('Missing user data');
            }

            const token = authTokenCookie.split('=')[1];
            const userData = JSON.parse(decodeURIComponent(userDataCookie.split('=')[1]));
            
            // Store the token securely
            await secureStorage.setSecureItem('token', token);
            await secureStorage.setSecureItem('userId', userData.id);
            
            // Set expiration time (24 hours from now)
            const expirationTime = Date.now() + (24 * 60 * 60 * 1000);
            await secureStorage.setSecureItem('tokenExpiration', expirationTime.toString());
            
            // Update auth store
            auth.set({
                user: userData,
                token,
                isLoading: false,
                error: null,
                isAuthenticated: true
            });
            
            // Ensure the auth state is updated before redirecting
            await new Promise(resolve => setTimeout(resolve, 100));
            
            // Redirect to chat
            goto('/chat', { replaceState: true });
        } catch (error) {
            console.error('OAuth callback error:', error);
            auth.setError(error instanceof Error ? error.message : 'Authentication failed');
            goto('/login', { replaceState: true });
        } finally {
            auth.setLoading(false);
        }
    });
</script>

<div class="flex items-center justify-center min-h-screen">
    <div class="text-center">
        <h2 class="text-2xl font-bold mb-4">Authenticating...</h2>
        {#if $auth.error}
            <p class="text-red-600">{$auth.error}</p>
        {/if}
    </div>
</div> 