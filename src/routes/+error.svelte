<script lang="ts">
  import { page } from '$app/stores'
	import Button from '$lib/components/Button.svelte';
	import { ArrowRight } from 'lucide-svelte';
</script>

<svelte:head>
  <title>{$page.status} - Mathéo Galuba</title>
</svelte:head>

<main class="grid min-h-full place-items-center px-6 py-24 sm:py-32 lg:px-8">
  <div class="container">
    <div class="text-center">
      <p class="text-base font-semibold text-ctp-mauve">{$page.status}</p>
      <h1 class="mt-4 text-3xl font-bold tracking-tight text-ctp-text sm:text-5xl">{$page.error?.message}</h1>
      <p class="mt-6 text-base leading-7 text-ctp-subtext0">Sorry, we couldn’t find the page you’re looking for.</p>
      <div>
        <pre class="mt-10 mx-auto w-fit antialiased text-left text-ctp-lavender">
  ／l、
（ﾟ､ ｡７
 |、ﾞ~ヽ
 じしf_, )ノ
        </pre>
      </div>
      <div class="mt-10 flex items-center justify-center gap-x-6">
        <Button link="/">
          <span>
            Go back home
          </span>
        </Button>
        <a href="/contact" class="flex justify-center items-center text-sm font-semibold text-ctp-text">Contact support&nbsp;<ArrowRight size="16"/></a>
      </div>
    </div>
  </div>
</main>
