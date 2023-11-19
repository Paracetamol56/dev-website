<script lang="ts">
	import { user, theme } from '$lib/stores';
	import { variants } from '@catppuccin/palette';
	import { createSelect, melt, type CreateSelectProps } from '@melt-ui/svelte';
	import axios from 'axios';
	import { AtSign, Check, ChevronDown, Palette, Save, User } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { fade } from 'svelte/transition';
	import { addToast } from '../+layout.svelte';

  let name: string;
  let nameError: string = "";
  let email: string;
  let emailError: string = "";
  let flavour: keyof typeof variants;

  onMount(() => {
    name = $user?.name || '';
    email = $user?.email || '';
    flavour = $theme as keyof typeof variants;
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
        flavour: flavour || $user?.flavour
      })
			.then((res) => {
				$user!.name = name;
        $user!.flavour = flavour || $user?.flavour;
        $theme = flavour || $user?.flavour;
			})
			.catch((err) => {
				console.error(err);
			});
	};

	const variantsClasses = ['ctp-latte', 'ctp-frappe', 'ctp-macchiato', 'ctp-mocha'];

  const onFlavourChange: CreateSelectProps['onSelectedChange'] = ({curr, next}) => {
    if (!next?.value) {
      return curr;
    }
    flavour = next.value as keyof typeof variants;
    return next;
  };

	const {
		elements: { trigger, menu, option, group, groupLabel, label },
		states: { selectedLabel, open },
		helpers: { isSelected }
	} = createSelect({
		forceVisible: true,
    defaultSelected: {
      value: $theme,
      label: $theme
    },
    onSelectedChange: onFlavourChange,
		positioning: {
			placement: 'bottom',
			fitViewport: true,
			sameWidth: true
		}
	});
</script>

<form class="grid grid-cols-1 gap-x-8 gap-y-6" on:submit={handleSave}>
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
			class="flex h-8 w-full items-center justify-between rounded-md bg-ctp-surface0 shadow-md shadow-ctp-crust px-3 focus:outline-none focus:ring-2 focus:ring-ctp-mauve"
			name="email"
			placeholder="Email"
			bind:value={email}
      on:blur={() => validateEmail()}
		/>
    <span class="text-left text-sm font-semibold text-ctp-red">{emailError}</span>
	</fieldset>
	<fieldset class="max-w-[15rem]">
		<label for="flavour" class="flex items-center gap-1 mb-2 text-sm font-semibold" use:melt={$label}><Palette size="16" /><span>Flavour</span></label>
		<button
			class="flex w-full items-center justify-between rounded-md bg-ctp-surface0 px-3 py-1
        shadow-md shadow-ctp-crust transition-opacity hover:opacity-90"
			use:melt={$trigger}
			aria-label="Food"
		>
			{$selectedLabel || 'Select a flavor'}
			<ChevronDown class="square-5" />
		</button>
		{#if $open}
			<div
				class="z-10 flex max-h-[300px] flex-col
    overflow-y-auto rounded-lg bg-ctp-base p-5 gap-3
    shadow-md shadow-ctp-crust focus:!ring-0"
				use:melt={$menu}
				transition:fade={{ duration: 150 }}
			>
				{#each Object.keys(variants) as variant, i}
					<div
						class="{variantsClasses[
							i
						]} bg-ctp-base relative cursor-pointer scroll-my-2 rounded-md py-2 pl-4 pr-4 text-ctp-text
                    {$isSelected(variant) ? 'bg-ctp-mauve/25 text-ctp-mauve' : ''}
                    hover:bg-ctp-mauve/25 hover:text-ctp-mauve"
						use:melt={$option({ value: variant, label: variant })}
					>
						{#if $isSelected(variant)}
							<div
								class="check absolute left-2 top-1/2 z-10 text-ctp-mauve transform -translate-y-1/2"
							>
								<Check class="square-4" />
							</div>
						{/if}

						<div class="pl-4">
							<p class="text-left capitalize font-medium">
								{variant}
							</p>
							<div class="flex w-full justify-between">
								<div class="square-2 rounded-full bg-ctp-rosewater" />
								<div class="square-2 rounded-full bg-ctp-flamingo" />
								<div class="square-2 rounded-full bg-ctp-pink" />
								<div class="square-2 rounded-full bg-ctp-mauve" />
								<div class="square-2 rounded-full bg-ctp-red" />
								<div class="square-2 rounded-full bg-ctp-maroon" />
								<div class="square-2 rounded-full bg-ctp-peach" />
								<div class="square-2 rounded-full bg-ctp-yellow" />
								<div class="square-2 rounded-full bg-ctp-green" />
								<div class="square-2 rounded-full bg-ctp-teal" />
								<div class="square-2 rounded-full bg-ctp-sky" />
								<div class="square-2 rounded-full bg-ctp-sapphire" />
								<div class="square-2 rounded-full bg-ctp-blue" />
								<div class="square-2 rounded-full bg-ctp-lavender" />
							</div>
						</div>
					</div>
				{/each}
				<small class="text-center">
					<img
						src={$theme === 'latte' ? '/img/catppuccin-light.png' : '/img/catppuccin-dark.png'}
						alt="Catppuccin logo"
						class="inline-block rounded-full square-4 mr-1"
					/>
					Powered by
					<a class="text-ctp-blue" href="https://github.com/catppuccin/catppuccin" target="_blank"
						>Catppuccin</a
					>
				</small>
			</div>
		{/if}
	</fieldset>
	<div class="flex justify-end">
		<button
			class="flex items-center gap-1 rounded-md bg-ctp-mauve px-3 py-1 font-semibold text-ctp-mantle shadow-md shadow-ctp-crust transition-opacity hover:opacity-80 active:opacity-60"
			type="submit"
    >
      <span>Save</span>
			<Save size="16" />
		</button>
	</div>
</form>
