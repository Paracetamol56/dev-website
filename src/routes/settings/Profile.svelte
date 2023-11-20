<script lang="ts">
	import { user } from '$lib/stores';
	import axios from 'axios';
	import { AtSign, Save, User } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { addToast } from '../+layout.svelte';
	import Button from '$lib/components/Button.svelte';

  let name: string;
  let nameError: string = "";
  let email: string;
  let emailError: string = "";

  onMount(() => {
    name = $user?.name || '';
    email = $user?.email || '';
  });

  const validateName = () => {
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
  const validateEmail = () => {
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
 
	const handleSave = async () => {
    if (!validateName() || !validateEmail()) {
      addToast({
        data: {
          title: 'Invalid form',
          description: 'Please check your inputs',
          color: 'bg-ctp-red'
        }
      });
      return;
    }

		axios
			.patch(`/api/user/${$user?.id}`, {
        name,
      })
			.then((res) => {
				$user!.name = name;
			})
			.catch((err) => {
				console.error(err);
			});
	};
</script>

<form class="grid gap-x-8 gap-y-6" on:submit={handleSave}>
	<fieldset class="max-w-[15rem]">
		<label for="name" class="flex items-center gap-1 mb-2 text-sm font-semibold"><User size="16"/><span>Name</span></label>
		<input
			type="text"
			id="name"
			class="flex h-8 w-full items-center justify-between rounded-md bg-ctp-surface0 shadow-md shadow-ctp-crust px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
			name="name"
			placeholder="Name"
			bind:value={name}
      on:blur={() => validateName()}
		/>
    <span class="text-left text-sm font-semibold text-ctp-red">{nameError}</span>
	</fieldset>
	<fieldset class="max-w-[15rem]">
		<label for="email" class="flex items-center gap-1 mb-2 text-sm font-semibold"><AtSign size="16"/><span>Email</span></label>
		<input
			type="email"
			disabled
			id="email"
			class="opacity-75 flex h-8 w-full items-center justify-between rounded-md bg-ctp-surface0 shadow-md shadow-ctp-crust px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
			name="email"
			placeholder="Email"
			bind:value={email}
      on:blur={() => validateEmail()}
		/>
    <span class="text-left text-sm font-semibold text-ctp-red">{emailError}</span>
	</fieldset>
	<div class="flex justify-start">
		<Button
			type="submit"
    >
      <span>Save</span>
			<Save size="16" />
		</Button>
	</div>
</form>
