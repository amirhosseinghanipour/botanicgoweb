<script lang="ts">
  import { store as websocketStore, isConnected } from "$lib/stores/websocket";
  import { page } from "$app/stores";

  let messageText = "";

  function handleSend() {
    if (messageText.trim() && $isConnected) {
      const sessionId = $page.params.sessionId;
      if (sessionId) {
        websocketStore.sendMessage(messageText, sessionId);
        messageText = "";
      }
    }
  }
</script>

<div class="relative">
  <textarea
    bind:value={messageText}
    on:keydown={(e) => {
      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault();
        handleSend();
      }
    }}
    placeholder="Type your message..."
    class="w-full resize-none rounded-lg border border-gray-300 bg-gray-100 p-3 pr-20 text-sm focus:outline-none dark:border-gray-600 dark:bg-gray-800"
    disabled={!$isConnected}
  />
  <button
    on:click={handleSend}
    disabled={!$isConnected || !messageText.trim()}
    class="absolute bottom-2 right-2 rounded-md bg-blue-500 px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
  >
    Send
  </button>
</div>
