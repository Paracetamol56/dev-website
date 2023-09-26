<script lang="ts">
  import { page } from '$app/stores';
	import { ArrowRight } from 'lucide-svelte';
</script>

<svelte:head>
  <title>{$page.error.message||'Error'} - Mathéo Galuba</title>
</svelte:head>

<main class="grid min-h-full place-items-center px-6 py-24 sm:py-32 lg:px-8">
  <div class="container">
    <div class="text-center">
      <p class="text-base font-semibold text-ctp-mauve">404</p>
      <h1 class="mt-4 text-3xl font-bold tracking-tight text-ctp-text sm:text-5xl">Page not found</h1>
      <p class="mt-6 text-base leading-7 text-ctp-subtext0">Sorry, we couldn’t find the page you’re looking for.</p>
      <div>
        <pre class="mt-10 mx-auto w-fit antialiased text-left text-ctp-subtext0">
  ／l、
（ﾟ､ ｡７
 |、ﾞ~ヽ
 じしf_, )ノ
        </pre>
      </div>
      <div class="mt-10 flex items-center justify-center gap-x-6">
        <a href="/" class="rounded-md bg-ctp-mauve px-3 py-1 font-medium text-ctp-surface0 hover:opacity-75 active:opacity-50 transition-opacity">Go back home</a>
        <a href="/contact" class="flex justify-center items-center text-sm font-semibold text-ctp-text">Contact support&nbsp;<ArrowRight size="16"/></a>
      </div>
    </div>
  </div>
</main>
