<script>
  import '../app.postcss';
  import { createDialog, melt } from '@melt-ui/svelte';
  import { Send, X } from 'lucide-svelte';

  const {
    elements: {
      trigger,
      overlay,
      content,
      title,
      description,
      close,
      portalled,
    },
    states: { open },
  } = createDialog({
    forceVisible: true,
  });
</script>

<!-- Button -->
<button
  class="rounded-md bg-ctp-mauve px-3 py-1 font-medium text-ctp-surface0 hover:opacity-75 active:opacity-50 transition-opacity"
  use:melt={$trigger}
>
  Login
</button>

<!-- Dialog -->
<div use:melt={$portalled}>
  {#if $open}
    <div use:melt={$overlay} class="fixed inset-0 z-50 bg-black/50" />
    <div
      class="fixed left-[50%] top-[50%] z-50 max-h-[85vh] w-[90vw]
            max-w-[450px] translate-x-[-50%] translate-y-[-50%] rounded-xl bg-ctp-base
            p-6 shadow-md"
      use:melt={$content}
    >
      <h2 use:melt={$title} class="m-0 text-lg font-medium text-ctp-text">
        Login
      </h2>
      <p use:melt={$description} class="mb-5 mt-2 leading-normal text-ctp-text">
        This is a passwordless login. We will send you a magic link to your email.
      </p>

      <form action="">
        <fieldset class="mb-4 flex items-center gap-5">
          <label class="w-[90px] text-right text-ctp-text" for="name">
            Email
          </label>
          <input
            class="inline-flex h-8 w-full flex-1 items-center justify-center
                      rounded-sm border border-solid px-3 leading-none text-ctp-text"
            id="name"
            type="email"
            required
          />
        </fieldset>
        <div class="mt-6 flex justify-end gap-4">
          <button
          class="flex justify-center items-center rounded-md bg-ctp-mauve px-3 py-1 font-medium
                text-ctp-surface0 hover:opacity-75 active:opacity-50 transition-opacity"
          type="submit"
          >
            Send the magic link&nbsp;<Send size="16" />
          </button>
        </div>
      </form>
      <button
        use:melt={$close}
        aria-label="close"
        class="absolute right-4 top-4 inline-flex h-6 w-6 appearance-none
                items-center justify-center rounded-full p-1 text-base
                hover:bg-ctp-mauve hover:text-ctp-base transition-colors"
      >
        <X class="square-4" />
      </button>
    </div>
  {/if}
</div>