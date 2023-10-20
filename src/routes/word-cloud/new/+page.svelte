<script lang="ts">
	import { user } from "$lib/stores";
	import { Send } from "lucide-svelte";
	import LoginDialog from "../../LoginDialog.svelte";
	import { addToast } from "../../+layout.svelte";
	import axios from "axios";
	import { goto } from "$app/navigation";

  let name: string = "";
  let nameError: string = "";
  let description: string = "";
  let descriptionError: string = "";

  const validateName = (name: string) => {
    name = name.trim();
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

  const validateDescription = (description: string) => {
    description = description.trim();
    if (description.length === 0) {
      return true;
    }
    if (description.length < 10) {
      descriptionError = "Description must be at least 10 characters long";
      return false;
    }
    if (description.length > 1000) {
      descriptionError = "Description must be less than 1000 characters long";
      return false;
    }
    descriptionError = "";
    return true;
  };

  const handleSubmit = (e: Event) => {
    e.preventDefault();
    if (!validateName(name) || !validateDescription(description)) {
      addToast({data:{
        title: "Error",
        description: "Please check your inputs",
        color: "bg-ctp-red",
      }});
      return;
    }
    axios.post("/api/word-cloud", {
      name,
      description,
    }).then((res) => {
      addToast({data:{
        title: "Success",
        description: "Word cloud created",
        color: "bg-ctp-green",
      }});
      goto(`/word-cloud/${res.data.id}`);
    }).catch((err) => {
      addToast({data:{
        title: "Error",
        description: "An error occured while creating the session",
        color: "bg-ctp-red",
      }});
    });
  };

</script>

<svelte:head>
	<title>World cloud - Mathéo Galuba</title>
</svelte:head>

<section class="container mx-auto mb-32">
	<hgroup>
		<h1 class="mb-8 text-6xl font-bold text-center">
			<span class="text-transparent bg-clip-text bg-gradient-to-r from-ctp-mauve to-ctp-lavender">Word cloud</span>
		</h1>
	</hgroup>
</section>

<section class="container mx-auto">
  <h2 class="mb-8 text-2xl font-semibold text-center">
    Create a new session
  </h2>
  {#if $user === null}
    <div class="mb-8 flex flex-col items-center gap-4">
      <p class="text-lg text-center text-ctp-red font-semibold">
        You must be logged in to create a new session
      </p>
      <LoginDialog />
    </div>
  {/if}
  <form class="mx-auto max-w-xl" on:submit={handleSubmit}>
    <div class="flex flex-col gap-y-6">
      <fieldset class="w-full">
        <label for="name" class="mb-2 text-sm font-semibold">
          Name
        </label>
        <input
          id="name"
          type="text"
          name="name"
          class="flex h-8 w-full items-center justify-between rounded-md bg-ctp-surface0
                px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
          bind:value={name}
          on:blur={() => validateName(name)}
          disabled={$user === null}
        />
        <p class="text-left text-sm font-semibold text-ctp-red">{nameError}</p>
      </fieldset>
      <fieldset class="sm:col-span-2">
        <label for="description" class="mb-2 text-sm font-semibold">
          Description
        </label>
        <textarea
          id="description"
          name="description"
          class="flex h-32 w-full items-center justify-between rounded-md bg-ctp-surface0
                px-3 py-2 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
          bind:value={description}
          on:blur={() => validateDescription(description)}
          disabled={$user === null}
        />
        <p class="text-left text-sm font-semibold text-ctp-red">{descriptionError}</p>
      </fieldset>
      <div class="sm:col-span-2 flex justify-end">
        <button
          class="flex justify-center items-center rounded-md bg-ctp-mauve px-3 py-1 font-medium
                text-ctp-surface0 hover:opacity-75 active:opacity-50 transition-opacity"
          type="submit"
        >
          Create&nbsp;<Send size="16" />
        </button>
      </div>
    </div>
  </form>
</section>
