<script lang="ts">
    import { auth } from '$lib/stores/auth';
    import { goto } from '$app/navigation';

    let handleLogout = async () => {
        try {
            await auth.clearAuth();
            goto('/auth/login');
        } catch (error) {
            console.error('Error during logout:', error);
        }
    };
</script>

<div class="flex items-center gap-4">
    {#if $auth.isLoading}
        <div class="flex items-center gap-2 text-gray-600">
            <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-gray-600"></div>
            <span class="text-sm">Please wait...</span>
        </div>
    {:else if $auth.isAuthenticated}
        <div class="flex items-center gap-4">
            {#if $auth.user?.avatarUrl}
                <img 
                    src={$auth.user.avatarUrl} 
                    alt="Profile" 
                    class="w-8 h-8 rounded-full"
                />
            {/if}
            <button
                on:click={handleLogout}
                class="text-sm text-gray-600 hover:text-gray-900 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                disabled={$auth.isLoading}
            >
                Logout
            </button>
        </div>
    {:else}
        <a
            href="/auth/login"
            class="text-sm text-gray-600 hover:text-gray-900 transition-colors"
        >
            Login
        </a>
    {/if}
</div> 