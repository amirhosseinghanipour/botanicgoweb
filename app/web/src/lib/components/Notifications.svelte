<script lang="ts">
  import { notifications } from '$lib/stores/notifications';
  import { fade, fly } from 'svelte/transition';
</script>

<div class="fixed top-4 right-4 z-50 flex flex-col gap-2">
  {#each $notifications as notification (notification.id)}
    <div
      class="p-4 rounded-lg shadow-lg max-w-md"
      class:bg-green-100={notification.type === 'success'}
      class:text-green-800={notification.type === 'success'}
      class:bg-red-100={notification.type === 'error'}
      class:text-red-800={notification.type === 'error'}
      class:bg-blue-100={notification.type === 'info'}
      class:text-blue-800={notification.type === 'info'}
      class:bg-yellow-100={notification.type === 'warning'}
      class:text-yellow-800={notification.type === 'warning'}
      transition:fly={{ y: -20, duration: 300 }}
    >
      <div class="flex justify-between items-start">
        <p class="text-sm font-medium">{notification.message}</p>
        <button
          class="ml-4 text-current opacity-50 hover:opacity-100"
          on:click={() => notifications.remove(notification.id)}
        >
          ×
        </button>
      </div>
    </div>
  {/each}
</div> 