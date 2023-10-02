<script lang="ts">
	import { createSwitch, melt } from "@melt-ui/svelte";
	import { Send } from "lucide-svelte";


  let email: string = "";
  let name: string = "";
  let message: string = "";

  let emailError: string = "";
  let nameError: string = "";
  let messageError: string = "";

  const {
    elements: { root, input },
  } = createSwitch();
</script>

<svelte:head>
	<title>Contact - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto my-32">
	<h1 class="mb-8 text-6xl font-bold text-center">
		<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">Contact</span>
	</h1>
  <form class="mx-auto max-w-xl" method="post">
    <div class="grid grid-cols-1 gap-x-8 gap-y-6 sm:grid-cols-2">
      <fieldset class="w-full">
        <label for="hip" class="mb-2 text-sm font-semibold">
          Email
        </label>
        <input
          id="email"
          type="email"
          name="email"
          class="flex h-8 w-full items-center justify-between rounded-md bg-ctp-text
                px-3 pr-12 text-ctp-base focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
          bind:value={email}
        />
        <p class="text-left text-sm font-semibold text-ctp-red">{emailError}</p>
      </fieldset>
      <fieldset class="w-full">
        <label for="hip" class="mb-2 text-sm font-semibold">
          Name
        </label>
        <input
          id="name"
          type="text"
          name="name"
          class="flex h-8 w-full items-center justify-between rounded-md bg-ctp-text
                px-3 text-ctp-base focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
          bind:value={name}
        />
        <p class="text-left text-sm font-semibold text-ctp-red">{nameError}</p>
      </fieldset>
      <fieldset class="sm:col-span-2">
        <label for="hip" class="mb-2 text-sm font-semibold">
          Message
        </label>
        <textarea
          id="message"
          name="message"
          class="flex h-32 w-full items-center justify-between rounded-md bg-ctp-text
                px-3 py-2 text-ctp-base focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
          bind:value={message}
        />
        <p class="text-left text-sm font-semibold text-ctp-red">{messageError}</p>
      </fieldset>
      <fieldset class="sm:col-span-2">
        <div class="flex items-center">
          <button
            use:melt={$root}
            class="switch relative h-5 w-[2.5rem] cursor-default rounded-full bg-ctp-crust transition-colors data-[state=checked]:bg-ctp-mauve"
            id="privacy"
          >
            <span class="thumb block rounded-full bg-white transition" />
          </button>
          <input use:melt={$input} />
          <label
            class="pl-4 leading-none text-sm font-semibold"
            for="privacy"
          >
            By checking this box, you agree to share your data with me.
          </label>
        </div>
      </fieldset>
      <div class="sm:col-span-2 flex justify-end">
        <button
          class="flex justify-center items-center rounded-md bg-ctp-mauve px-3 py-1 font-medium
                text-ctp-surface0 hover:opacity-75 active:opacity-50 transition-opacity"
          type="submit"
        >
          Send&nbsp;<Send size="16" />
        </button>
      </div>
    </div>
  </form>
</section>

<style>
  .switch {
    --w: 2.25rem;
    --padding: 0.125rem;
    width: var(--w);
  }
 
  .thumb {
    --size: 1rem;
    width: var(--size);
    height: var(--size);
    transform: translateX(var(--padding));
  }
 
  :global([data-state='checked']) .thumb {
    transform: translateX(calc(var(--w) - var(--size) - var(--padding)));
  }
</style>