<script lang="ts">
	import { createTabs, melt } from '@melt-ui/svelte';
	import { cubicInOut } from 'svelte/easing';
	import { crossfade } from 'svelte/transition';
	import { Highlight } from "svelte-highlight";
	import { python, cpp, javascript, rust } from 'svelte-highlight/languages';

	const {
		elements: { root, list, content, trigger },
		states: { value }
	} = createTabs({
		defaultValue: 'tab-1'
	});

  const [send, receive] = crossfade({
    duration: 250,
    easing: cubicInOut,
  });

	export let code = [
		{
			id: 'tab-1',
			title: 'Python',
      lang: python,
			code: 'print("Hello World!")'
		},
		{
			id: 'tab-2',
			title: 'JavaScript',
      lang: javascript,
			code: 'console.log("Hello World!")'
		},
		{
			id: 'tab-3',
			title: 'Rust',
      lang: rust,
			code: 'fn main() { println!("Hello, world!"); }'
		},
		{
			id: 'tab-4',
			title: 'C++',
      lang: cpp,
			code: 'std::cout << "Hello World!" << std::endl;'
		}
	];
</script>

<div
  use:melt={$root}
  class={'flex xl:mx-32 2xl:mx-64 flex-col overflow-hidden rounded-md shadow-md shadow-ctp-crust data-[orientation=vertical]:flex-row'}
>
  <div
    use:melt={$list}
    class="flex shrink-0 overflow-x-auto bg-ctp-crust
    data-[orientation=vertical]:flex-col data-[orientation=vertical]:border-r"
    aria-label="Tabs"
  >
    {#each code as triggerItem}
      <button
        use:melt={$trigger(triggerItem.id)}
        class="relative flex items-center justify-center cursor-default select-none
              rounded-none font-medium leading-1
              flex-1 h-12 px-2 focus:relative { $value === triggerItem.id ? 'bg-ctp-mantle' : 'bg-ctp-crust'}"
      >
        {triggerItem.title}
        {#if $value === triggerItem.id}
          <div
            in:send={{ key: 'trigger' }}
            out:receive={{ key: 'trigger' }}
            class="absolute bottom-1 left-1/2 h-1 w-6 -translate-x-1/2 rounded-full bg-ctp-mauve"
          />
        {/if}
      </button>
    {/each}
  </div>
  {#each code as tab}
    <div use:melt={$content(tab.id)} class="grow bg-ctp-mantle p-5">
      <Highlight language={tab.lang} code={tab.code} />
    </div>
  {/each}
</div>
