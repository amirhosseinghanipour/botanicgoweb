<script lang="ts">
  import type { Message } from "$lib/types";
  import { auth } from "$lib/stores/auth";

  export let message: Message;

  $: isUser = message.role === "user" || message.userId === $auth.user?.id;
</script>

<div class="flex items-start gap-4" class:justify-end={isUser}>
  {#if !isUser}
    <span
      class="relative flex h-10 w-10 shrink-0 overflow-hidden rounded-full items-center justify-center bg-gray-200 dark:bg-gray-700"
    >
      AI
    </span>
  {/if}
  <div
    class="max-w-[75%] rounded-lg p-3 text-sm"
    class:bg-blue-500={isUser}
    class:text-white={isUser}
    class:bg-gray-200={!isUser}
    class:dark:bg-gray-700={!isUser}
    class:dark:text-gray-200={!isUser}
  >
    <p>{message.content}</p>
    <div
      class="mt-1 text-xs"
      class:text-gray-300={isUser}
      class:text-gray-500={!isUser}
    >
      {new Date(message.created_at).toLocaleTimeString()}
    </div>
  </div>
  {#if isUser}
    <span
      class="relative flex h-10 w-10 shrink-0 overflow-hidden rounded-full items-center justify-center bg-blue-500 text-white"
    >
      You
    </span>
  {/if}
</div>
