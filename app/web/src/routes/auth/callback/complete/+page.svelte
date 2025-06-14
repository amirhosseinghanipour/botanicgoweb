<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { auth } from '$lib/stores/auth';
    import { secureStorage } from '$lib/utils/secureStorage';
    import { API_URL } from '$lib/config';
    import { api } from '$lib/api/client';

    let error: string | null = null;
    let loading = true;

    onMount(async () => {
        try {
            // Get the encoded data from URL
            const urlParams = new URLSearchParams(window.location.search);
            const encodedData = urlParams.get('data');

            console.log('URL params:', Object.fromEntries(urlParams.entries()));

            if (!encodedData) {
                console.error('Missing authentication data in URL');
                goto('/login?error=missing_auth_data');
                return;
            }

            try {
                // Decode the data
                const decodedData = atob(encodedData);
                console.log('Decoded data:', decodedData);
                
                const authData = JSON.parse(decodedData);
                console.log('Parsed auth data:', authData);

                if (!authData.token || !authData.user) {
                    console.error('Invalid auth data format:', authData);
                    throw new Error('Invalid authentication data format');
                }

                // Store the token in secureStorage with expiration
                const expirationTime = Date.now() + (30 * 24 * 60 * 60 * 1000); // 30 days
                await secureStorage.setSecureItem('token', authData.token);
                await secureStorage.setSecureItem('tokenExpiration', expirationTime.toString());
                
                // Store user data in localStorage
                localStorage.setItem('user', JSON.stringify(authData.user));

                // Initialize the API client with the new token
                await api.initializeToken(authData.token, expirationTime);

                // Create a new session if one doesn't exist
                try {
                    const response = await fetch(`${API_URL}/api/chat/sessions`, {
                        method: 'POST',
                        headers: {
                            'Authorization': `Bearer ${authData.token}`,
                            'Content-Type': 'application/json'
                        },
                        body: JSON.stringify({
                            title: 'New Chat'
                        }),
                        credentials: 'include'
                    });

                    if (response.ok) {
                        const data = await response.json();
                        if (data.session && data.session.id) {
                            await secureStorage.setSecureItem('sessionId', data.session.id);
                            await api.initializeToken(authData.token, expirationTime, data.session.id);
                        }
                    } else {
                        console.error('Failed to create session:', await response.text());
                    }
                } catch (error) {
                    console.error('Failed to create session:', error);
                }

                // Read back and log to verify storage
                const testToken = await secureStorage.getSecureItem('token');
                const testExpiration = await secureStorage.getSecureItem('tokenExpiration');
                const testSessionId = await secureStorage.getSecureItem('sessionId');
                const testUser = localStorage.getItem('user');
                console.log('Storage verification:', {
                    token: testToken ? 'present' : 'missing',
                    expiration: testExpiration ? 'present' : 'missing',
                    sessionId: testSessionId ? 'present' : 'missing',
                    user: testUser ? 'present' : 'missing'
                });

                // Verify token with backend before proceeding
                try {
                    const response = await fetch(`${API_URL}/api/auth/verify`, {
                        method: 'POST',
                        headers: {
                            'Authorization': `Bearer ${authData.token}`,
                            'Content-Type': 'application/json'
                        },
                        credentials: 'include'
                    });

                    if (!response.ok) {
                        throw new Error('Token verification failed');
                    }

                    // Wait a tick to ensure storage is flushed
                    await new Promise(res => setTimeout(res, 100));

                    // Now initialize the auth store
                    await auth.initializeAuth();

                    // Redirect to chat page
                    goto('/chat');
                } catch (error) {
                    console.error('Token verification failed:', error);
                    goto('/login?error=token_verification_failed');
                }
            } catch (error) {
                console.error('Error processing authentication data:', error);
                goto('/login?error=invalid_auth_data');
            }
        } catch (error) {
            console.error('Error in auth callback:', error);
            goto('/login?error=auth_callback_failed');
        } finally {
            loading = false;
        }
    });
</script>

<div class="min-h-screen flex items-center justify-center">
    <div class="max-w-md w-full mx-auto p-8">
        {#if loading}
            <div class="flex flex-col items-center space-y-6">
                <div class="text-center">
                    <h1 class="text-3xl font-bold tracking-tight text-gray-900 dark:text-white mb-2">
                        Completing Authentication
                    </h1>
                </div>
                <div class="flex items-center justify-center">
                    <svg
                        class="h-8 w-8 animate-spin text-gray-900 dark:text-gray-50"
                        viewBox="0 0 24 24"
                        fill="none"
                        xmlns="http://www.w3.org/2000/svg"
                    >
                        <path d="M12 4.75V6.25" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                        <path
                            d="M17.1475 6.8525L16.0625 7.9375"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                        />
                        <path d="M19.25 12H17.75" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                        <path
                            d="M17.1475 17.1475L16.0625 16.0625"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                        />
                        <path d="M12 17.75V19.25" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                        <path
                            d="M6.8525 17.1475L7.9375 16.0625"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                        />
                        <path d="M4.75 12H6.25" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                        <path
                            d="M6.8525 6.8525L7.9375 7.9375"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                        />
                    </svg>
                </div>
            </div>
        {:else if error}
            <div class="text-center space-y-6">
                <div class="space-y-2">
                    <h1 class="text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
                        Authentication Failed
                    </h1>
                    <p class="text-red-600 dark:text-red-400">
                        {error}
                    </p>
                </div>
                <button
                    class="inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-white bg-black dark:text-black dark:bg-white rounded-md hover:opacity-90 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-black dark:focus:ring-white transition-all duration-200"
                    on:click={() => goto('/login')}
                >
                    Return to Login
                </button>
            </div>
        {/if}
    </div>
</div> 