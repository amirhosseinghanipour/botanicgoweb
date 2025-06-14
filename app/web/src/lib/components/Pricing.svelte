<script>
  import { icons } from "$lib/icons.js";
</script>

<div class="flex flex-col items-center text-center">
  <h1 class="text-4xl md:text-5xl font-bold tracking-tighter mb-6">
    Free, Open, and Community-Driven
  </h1>
  <p class="max-w-2xl text-gray-500 dark:text-gray-400 mb-12">
    Botanic is a passion project committed to providing powerful AI tools freely
    to the developer community. We believe in the power of open source.
  </p>

  <div
    class="w-full max-w-2xl bg-gray-50 dark:bg-neutral-900/50 border border-gray-200 dark:border-gray-800 rounded-2xl p-8 md:p-12"
  >
    <div class="w-16 h-16 mx-auto mb-6 text-pink-500">
      {@html icons.heart}
    </div>
    <h2 class="text-2xl font-bold mb-4">Support the Project</h2>
    <p class="text-gray-600 dark:text-gray-400 mb-8">
      Your contributions help cover server costs, support development, and allow
      us to continue building and improving Botanic for everyone. If you find
      this tool valuable, please consider making a donation.
    </p>

    <div class="grid sm:grid-cols-2 gap-4">
      <a
        href="https://github.com/sponsors"
        target="_blank"
        rel="noopener noreferrer"
        class="group flex items-center justify-center gap-3 w-full py-3 px-4 text-white dark:text-black bg-black dark:bg-white rounded-lg font-semibold transition-transform hover:scale-105"
      >
        {@html icons.github}
        <span>Sponsor on GitHub</span>
      </a>
      <a
        href="https://ko-fi.com"
        target="_blank"
        rel="noopener noreferrer"
        class="group flex items-center justify-center gap-3 w-full py-3 px-4 bg-[#FF5E5B] text-white rounded-lg font-semibold transition-transform hover:scale-105"
      >
        <span class="w-6 h-6">{@html icons.heart}</span>
        <span>Buy us a Coffee</span>
      </a>
    </div>
  </div>

  <div class="mt-16 text-center">
    <h3 class="text-xl font-bold mb-4">Want to contribute with code?</h3>
    <p class="text-gray-500 dark:text-gray-400 mb-6 max-w-lg">
      We welcome developers of all skill levels to contribute. Check out our
      repository, report issues, or submit a pull request!
    </p>
    <a
      href="#"
      target="_blank"
      rel="noopener noreferrer"
      class="inline-block px-8 py-3 bg-gray-100 dark:bg-gray-800 border border-gray-300 dark:border-gray-700 rounded-lg font-semibold hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
    >
      View on GitHub
    </a>
  </div>
</div>
