<script lang="ts">
	export let data
</script>

<svelte:head>
	<title>{data.meta.title}</title>
	<meta property="og:type" content="article" />
	<meta property="og:title" content={data.meta.title} />
</svelte:head>

<article>
  <section class="container mx-auto my-32">
    <hgroup>
      <h1 class="mb-8 text-6xl font-bold text-center">
        <span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">
          {data.meta.title}
        </span>
      </h1>
    </hgroup>

    <div class="flex gap-2 justify-center">
      {#each data.meta.categories as category}
        <span class="font-semibold text-ctp-lavender">
          &num;{category}
        </span>
      {/each}
    </div>
  </section>

	<div class="container mx-auto content">
		<svelte:component this={data.content} />
	</div>
</article>
