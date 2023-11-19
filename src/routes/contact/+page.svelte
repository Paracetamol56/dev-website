<script lang="ts">
	import { createSwitch, melt, type CreateSwitchProps } from "@melt-ui/svelte";
	import axios from "axios";
	import { Send } from "lucide-svelte";
	import { addToast } from "../+layout.svelte";
	import { user } from "$lib/stores";
	import Button from "$lib/components/Button.svelte";

  let email: string = "";
  let name: string = "";
  let message: string = "";

  let emailError: string = "";
  let nameError: string = "";
  let messageError: string = "";
  let privacyError: string = "";

  const validateEmail = (email: string) => {
    if (email.length === 0) {
      emailError = "Email is required";
      return false;
    }
    if (!/\S+@\S+\.\S+/.test(email)) {
      emailError = "Email address is invalid";
      return false;
    }
    emailError = "";
    return true;
  };
  const validateName = (name: string) => {
    if (name.length === 0) {
      nameError = "Name is required";
      return false;
    }
    if (name.length < 3) {
      nameError = "Name must be at least 3 characters long";
      return false;
    }
    if (name.length > 100) {
      nameError = "Name must be less than 100 characters long";
      return false;
    }
    nameError = "";
    return true;
  };
  const validateMessage = (message: string) => {
    if (message.length === 0) {
      messageError = "Message is required";
      return false;
    }
    if (message.length < 10) {
      messageError = "Message must be at least 10 characters long";
      return false;
    }
    if (message.length > 1000) {
      messageError = "Message must be less than 1000 characters long";
      return false;
    }
    messageError = "";
    return true;
  };
  const validatePrivacy: CreateSwitchProps['onCheckedChange'] = ({curr, next}) => {
    if (!next) {
      privacyError = "This field is required";
    } else {
      privacyError = "";
    }
    return next;
  };

  const handleSubmit = (e: Event) => {
    e.preventDefault();
    if (!validateEmail(email) || !validateName(name) || !validateMessage(message) || !$checked) {
      addToast({
        data: {
          title: "Invalid form",
          description: "Please check your inputs",
          color: "bg-ctp-red"
        }
      });
      return;
    }

    axios.post('/api/contact', {
      userId: $user?.id ?? null,
      email,
      name,
      message,
    }).then(() => {
      addToast({
        data: {
          title: "Message sent",
          description: "I will get back to you as soon as possible",
          color: "bg-ctp-green"
        }
      });
      email = "";
      name = "";
      message = "";
    }).catch(() => {
      addToast({
        data: {
          title: "Error",
          description: "An error occured while sending your message",
          color: "bg-ctp-red"
        }
      });
    });
  };

  const {
    elements: { root, input },
    states: { checked },
  } = createSwitch({
    name: "privacy",
    defaultChecked: false,
    onCheckedChange: validatePrivacy,
  });
</script>

<svelte:head>
	<title>Contact - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto">
	<h1 class="mb-8 text-6xl font-bold text-center">
		<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">Contact</span>
	</h1>
  <form class="mx-auto max-w-xl" on:submit={handleSubmit}>
    <div class="grid grid-cols-1 gap-x-8 gap-y-6 sm:grid-cols-2">
      <fieldset class="w-full">
        <label for="email" class="mb-2 text-sm font-semibold">
          Email
        </label>
        <input
          id="email"
          type="email"
          name="email"
          class="flex h-8 w-full items-center justify-between rounded-md bg-ctp-surface0
                shadow-md shadow-ctp-crust px-3 pr-12 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
          bind:value={email}
          on:blur={() => validateEmail(email)}
        />
        <p class="text-left text-sm font-semibold text-ctp-red">{emailError}</p>
      </fieldset>
      <fieldset class="w-full">
        <label for="name" class="mb-2 text-sm font-semibold">
          Name
        </label>
        <input
          id="name"
          type="text"
          name="name"
          class="flex h-8 w-full items-center justify-between rounded-md bg-ctp-surface0
                shadow-md shadow-ctp-crust px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
          bind:value={name}
          on:blur={() => validateName(name)}
        />
        <p class="text-left text-sm font-semibold text-ctp-red">{nameError}</p>
      </fieldset>
      <fieldset class="sm:col-span-2">
        <label for="message" class="mb-2 text-sm font-semibold">
          Message
        </label>
        <textarea
          id="message"
          name="message"
          class="flex h-32 w-full items-center justify-between rounded-md bg-ctp-surface0
                shadow-md shadow-ctp-crust px-3 py-2 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
          bind:value={message}
          on:blur={() => validateMessage(message)}
        />
        <p class="text-left text-sm font-semibold text-ctp-red">{messageError}</p>
      </fieldset>
      <fieldset class="sm:col-span-2">
        <div class="flex items-center justify-start">
          <button
            use:melt={$root}
            class="switch relative h-5 min-w-[2.25rem] cursor-default rounded-full bg-ctp-surface0
                  shadow-md shadow-ctp-crust transition-colors data-[state=checked]:bg-ctp-mauve"
            id="privacy"
          >
            <span class="thumb block rounded-full bg-ctp-text transition" />
          </button>
          <input use:melt={$input} />
          <label
            class="pl-4 leading-none text-sm font-semibold"
            for="privacy"
          >
            By checking this box, you agree to share your data with me.
          </label>
        </div>
        <p class="text-left text-sm font-semibold text-ctp-red">{privacyError}</p>
      </fieldset>
      <div class="sm:col-span-2 flex justify-end">
        <Button type="submit">
          <span>Send</span>
          <Send size="16" />
        </Button>
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