<script lang="ts">
	import axios from 'axios';
	import { Send } from 'lucide-svelte';
	import { writable } from 'svelte/store';
	import { user } from '$lib/store';
	import { addToast } from '$lib/toast';
	import Button from '$lib/components/ui/Button.svelte';
	import Heading from '$lib/components/ui/Heading.svelte';
	import Hero from '$lib/components/ui/Hero.svelte';
	import Switch from '$lib/components/ui/Switch.svelte';
	import Text from '$lib/components/ui/Text.svelte';
	import Textarea from '$lib/components/ui/Textarea.svelte';
	import TextInput from '$lib/components/ui/TextInput.svelte';
	import { TextColor, TextSize } from '$lib/components/ui/typography';

	const MESSAGE_MAX = 1000;

	const email = writable('');
	const name = writable('');
	const message = writable('');
	const privacy = writable(false);

	let emailError: string = '';
	let nameError: string = '';
	let messageError: string = '';
	let privacyError: string = '';
	let submitted: boolean = false;
	let sending: boolean = false;

	const validateEmail = (email: string) => {
		if (email.length === 0) {
			emailError = 'Email is required';
			return false;
		}
		if (!/\S+@\S+\.\S+/.test(email)) {
			emailError = 'Email address is invalid';
			return false;
		}
		emailError = '';
		return true;
	};
	const validateName = (name: string) => {
		if (name.length === 0) {
			nameError = 'Name is required';
			return false;
		}
		if (name.length < 2) {
			nameError = 'Name must be at least 2 characters long';
			return false;
		}
		if (name.length > 100) {
			nameError = 'Name must be less than 100 characters long';
			return false;
		}
		nameError = '';
		return true;
	};
	const validateMessage = (message: string) => {
		if (message.length === 0) {
			messageError = 'Message is required';
			return false;
		}
		if (message.length < 10) {
			messageError = 'Message must be at least 10 characters long';
			return false;
		}
		if (message.length > MESSAGE_MAX) {
			messageError = `Message must be less than ${MESSAGE_MAX} characters long`;
			return false;
		}
		messageError = '';
		return true;
	};

	$: privacyError = submitted && !$privacy ? 'This field is required' : '';

	const handleSubmit = () => {
		submitted = true;
		const valid = [validateEmail($email), validateName($name), validateMessage($message), $privacy];
		if (!valid.every(Boolean)) {
			addToast({
				data: {
					title: 'Invalid form',
					description: 'Please check your inputs',
					color: 'bg-ctp-red'
				}
			});
			return;
		}

		sending = true;
		axios
			.post('/api/contact', {
				userId: $user?.id ?? null,
				email: $email,
				name: $name,
				message: $message
			})
			.then(() => {
				addToast({
					data: {
						title: 'Message sent',
						description: 'I will get back to you as soon as possible',
						color: 'bg-ctp-green'
					}
				});
				email.set('');
				name.set('');
				message.set('');
				submitted = false;
			})
			.catch(() => {
				addToast({
					data: {
						title: 'Error',
						description: 'An error occured while sending your message',
						color: 'bg-ctp-red'
					}
				});
			})
			.finally(() => (sending = false));
	};
</script>

<svelte:head>
	<title>Contact - Mathéo Galuba</title>
</svelte:head>

<Hero
	class="mb-12"
	highlight="Contact"
	description="A question, a bug to report or a project to talk about? Send me a message."
/>

<section class="container mx-auto mb-16 max-w-2xl">
	<form
		class="rounded-md border-2 border-transparent bg-ctp-crust/50 p-6 shadow-md shadow-ctp-crust backdrop-blur-sm
			transition-colors focus-within:border-ctp-mauve sm:p-8"
		novalidate
		on:submit|preventDefault={handleSubmit}
	>
		<Heading level={2} size={TextSize.Xl2} class="mb-1">Send a message</Heading>
		<div class="grid grid-cols-1 gap-x-6 gap-y-5 sm:grid-cols-2">
			<TextInput
				value={name}
				label="Name"
				name="name"
				autocomplete="name"
				error={nameError}
				on:blur={() => validateName($name)}
			/>
			<TextInput
				value={email}
				type="email"
				label="Email"
				name="email"
				autocomplete="email"
				error={emailError}
				on:blur={() => validateEmail($email)}
			/>
			<div class="sm:col-span-2">
				<Textarea
					value={message}
					rows={8}
					label="Message"
					name="message"
					error={messageError}
					hint="{$message.length} / {MESSAGE_MAX}"
					on:blur={() => validateMessage($message)}
				/>
			</div>
			<div class="sm:col-span-2">
				<Switch
					checked={privacy}
					name="privacy"
					label="By checking this box, you agree to share your data with me."
				/>
				{#if privacyError}
					<Text size={TextSize.Sm} color={TextColor.Danger} class="mt-1 font-semibold">
						{privacyError}
					</Text>
				{/if}
			</div>
			<div class="flex justify-end sm:col-span-2">
				<Button type="submit" loading={sending} data-umami-event="contact">
					<span>Send</span>
					<Send size="16" />
				</Button>
			</div>
		</div>
	</form>
</section>
