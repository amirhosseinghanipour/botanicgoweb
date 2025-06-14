<script>
  import { icons } from "$lib/icons.js";
</script>

<div class="flex flex-col items-center text-center">
  <h1 class="text-4xl md:text-5xl font-bold tracking-tighter mb-6">
    Powerful Tools for Modern Developers
  </h1>
  <p class="max-w-2xl text-gray-500 dark:text-gray-400 mb-16">
    Botanic is more than just a chat interface. It's an open-source toolkit
    designed to supercharge your workflow with cutting-edge AI capabilities.
  </p>

  <div class="grid md:grid-cols-2 gap-12 items-center mb-24 w-full">
    <div class="text-left">
      <div class="w-12 h-12 mb-4 text-black dark:text-white">
        {@html icons.multiPanel}
      </div>
      <h3 class="text-2xl font-bold mb-4">Multi-Panel View</h3>
      <p class="text-gray-500 dark:text-gray-400">
        Don't settle for one answer. Compare and contrast responses from
        multiple AI models simultaneously in a clean, side-by-side interface.
        Find the most optimal, creative, or efficient solution by leveraging the
        unique strengths of each model.
      </p>
    </div>
    <div
      class="bg-gray-100 dark:bg-neutral-900/50 p-8 rounded-lg flex items-center justify-center"
    >
      <div
        class="w-full h-48 bg-gray-300 dark:bg-gray-700 rounded-md shadow-lg text-gray-500 dark:text-gray-400 flex items-center justify-center"
      >
        Placeholder Image
      </div>
    </div>
  </div>

  <div class="grid md:grid-cols-2 gap-12 items-center mb-24 w-full">
    <div
      class="bg-gray-100 dark:bg-neutral-900/50 p-8 rounded-lg order-last md:order-first flex items-center justify-center"
    >
      <div
        class="w-full h-48 bg-gray-300 dark:bg-gray-700 rounded-md shadow-lg text-gray-500 dark:text-gray-400 flex items-center justify-center"
      >
        Placeholder Image
      </div>
    </div>
    <div class="text-left">
      <div class="w-12 h-12 mb-4 text-black dark:text-white">
        {@html icons.deepResearch}
      </div>
      <h3 class="text-2xl font-bold mb-4">Deep Research Mode</h3>
      <p class="text-gray-500 dark:text-gray-400">
        Go beyond surface-level answers. Activate Deep Research mode to have the
        AI perform comprehensive analysis, gather data from multiple sources,
        and provide you with well-supported, in-depth explanations for complex
        topics.
      </p>
    </div>
  </div>

  <div class="grid md:grid-cols-3 gap-8 w-full">
    <div
      class="p-6 bg-gray-50 dark:bg-neutral-900/50 border border-gray-200 dark:border-gray-800 rounded-lg text-left"
    >
      <div class="w-10 h-10 mb-4 text-black dark:text-white">
        {@html icons.codeExport}
      </div>
      <h3 class="text-xl font-bold mb-2">Code Export</h3>
      <p class="text-gray-500 dark:text-gray-400 text-sm">
        Instantly export generated code snippets to your favorite IDE or save
        them as files, maintaining formatting and syntax highlighting.
      </p>
    </div>
    <div
      class="p-6 bg-gray-50 dark:bg-neutral-900/50 border border-gray-200 dark:border-gray-800 rounded-lg text-left"
    >
      <div class="w-10 h-10 mb-4 text-black dark:text-white">
        {@html icons.chatHistory}
      </div>
      <h3 class="text-xl font-bold mb-2">Persistent History</h3>
      <p class="text-gray-500 dark:text-gray-400 text-sm">
        Never lose a great idea. Your conversations are saved locally, allowing
        you to search and revisit past interactions with ease.
      </p>
    </div>
    <div
      class="p-6 bg-gray-50 dark:bg-neutral-900/50 border border-gray-200 dark:border-gray-800 rounded-lg text-left"
    >
      <div class="w-10 h-10 mb-4 text-black dark:text-white">
        {@html icons.devFocused}
      </div>
      <h3 class="text-xl font-bold mb-2">Developer Focused</h3>
      <p class="text-gray-500 dark:text-gray-400 text-sm">
        Built by developers, for developers. Botanic is designed to integrate
        seamlessly into your workflow with a focus on code and technical
        queries.
      </p>
    </div>
  </div>
</div>
