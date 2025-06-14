<script lang="ts">
  import ProfileSettings from '$lib/components/ProfileSettings.svelte';
  import SessionManager from '$lib/components/SessionManager.svelte';
  import { page } from '$app/stores';
</script>

<div class="container mx-auto px-4 py-8">
  <div class="max-w-4xl mx-auto space-y-8">
    <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Settings</h1>

    <div class="bg-white dark:bg-gray-800 shadow rounded-lg divide-y divide-gray-200 dark:divide-gray-700">
      <div class="p-6">
        <h2 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">Profile Settings</h2>
        <ProfileSettings />
      </div>

      <div class="p-6">
        <h2 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">Active Sessions</h2>
        <SessionManager />
      </div>
    </div>
  </div>
</div> 