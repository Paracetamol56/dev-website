<script lang="ts">
  import '../app.postcss';
  import { createDialog, melt } from '@melt-ui/svelte';
  import { Send, X } from 'lucide-svelte';
	import { addToast } from './+layout.svelte';
  import axios from 'axios';
	import { writable, type Writable } from 'svelte/store';
	import { fade, fly } from 'svelte/transition';

  const dialogOpen: Writable<boolean> = writable(false);
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
    open: dialogOpen,
  });

  let email = '';
  let emailError = '';

  function validateEmail(): boolean {
    if (email.length == 0) {
      emailError = 'Email is required';
      return false;
    }
    if (!email.match(/^[^\s@]+@[^\s@]+\.[^\s@]+$/)) {
      emailError = 'Invalid email';
      return false;
    }
    emailError = '';
    return true;
  }

  const handleSubmit = (event: Event) => {
    event.preventDefault();
    
    if (!validateEmail()) {
      addToast({
        data: {
          title: 'Invalid form',
          description: 'Please check your inputs',
          color: 'bg-ctp-red'
        }
      });
      return;
    }

    dialogOpen.set(false);

    // Call the route /auth/login with the email as a parameter with axios
    axios.post('/auth/login', { email })
      .then((response) => {
        if (response.status != 200) {
          console.error(response);
          addToast({
            data: {
              title: 'Error',
              description: 'An error occured, please try again later',
              color: 'bg-ctp-red'
            }
          });
          return;
        }
        addToast({
          data: {
            title: 'Magic link sent',
            description: 'Please check your inbox',
            color: 'bg-ctp-green'
          }
        });
      })
      .catch((error) => {
        console.error(error);
        addToast({
          data: {
            title: 'Error',
            description: 'An error occured, please try again later',
            color: 'bg-ctp-red'
          }
        });
      });
  }
</script>

<!-- Button -->
<button
  class="rounded-md bg-ctp-mauve px-3 py-1 font-medium text-ctp-mantle hover:opacity-75 active:opacity-50 transition-opacity"
  use:melt={$trigger}
>
  Login
</button>

<!-- Dialog -->
<div use:melt={$portalled}>
  {#if $open}
    <div use:melt={$overlay} class="fixed inset-0 z-10 bg-black/50" transition:fade={{duration: 200}} />
    <div
      class="fixed left-[50%] top-[50%] z-50 max-h-[85vh] w-[90vw]
            max-w-[450px] translate-x-[-50%] translate-y-[-50%] rounded-lg bg-ctp-base
            p-6 shadow-md"
      transition:fly={{ duration: 200, y: 10 }}
      use:melt={$content}
    >
      <h2 use:melt={$title} class="m-0 text-lg font-medium text-ctp-text">
        Login
      </h2>
      <p use:melt={$description} class="mb-5 mt-2 leading-normal text-ctp-text">
        This is a passwordless authentication. We only need your email to send you a magic link.
        <br>
        <strong>
          Don't forget to check your spam folder!
        </strong>
      </p>

      <form on:submit={handleSubmit}>
        <fieldset class="flex items-center gap-5">
          <label class="text-right text-ctp-text" for="name">
            Email
          </label>
          <input
            class="h-8 flex-1 rounded-md border border-solid px-3 leading-none text-ctp-base
                  focus:ring-2 focus:ring-ctp-mauve focus:outline-none {emailError != '' ? 'ring-2 ring-ctp-red' : ''}"
            id="email"
            bind:value={email}
            on:blur={validateEmail}
            type="email"
          />
        </fieldset>
        <p class="mb-4 text-right text-sm text-ctp-red">{emailError}</p>
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
