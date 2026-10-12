<script lang="ts">
	import {
		Accordion,
		Badge,
		Button,
		ButtonVariant,
		Card,
		Checkbox,
		Code,
		Dialog,
		Divider,
		DropdownMenu,
		Heading,
		Hero,
		IconButton,
		Link,
		Loader,
		NumberInput,
		Pagination,
		Popover,
		Progress,
		RadioGroup,
		Select,
		Slider,
		Switch,
		Tabs,
		Text,
		Textarea,
		TextInput,
		Tooltip,
		TextColor,
		TextSize,
		TextWeight,
		Tone,
		type MenuItem
	} from '$lib/components/ui';
	import CodeBlock from '$lib/components/CodeBlock.svelte';
	import ColorPicker from '$lib/components/ColorPicker.svelte';
	import FileUploader from '$lib/components/FileUploader.svelte';
	import GitHub from '$lib/components/GitHub.svelte';
	import Google from '$lib/components/Google.svelte';
	import MeltDatePicker from '$lib/components/MeltDatePicker.svelte';
	import MeltTags from '$lib/components/MeltTags.svelte';
	import MultiLangCodeBlock from '$lib/components/MultiLangCodeBlock.svelte';
	import PaletteWheel from '$lib/components/PaletteWheel.svelte';
	import ProviderIcon from '$lib/components/ProviderIcon.svelte';
	import Demo from './Demo.svelte';
	import Section from './Section.svelte';
	import { addToast } from '../+layout.svelte';
	import { user } from '$lib/store';
	import type { PaletteColor } from '$lib/colors';
	import { variants } from '@catppuccin/palette';
	import type { DateValue } from '@internationalized/date';
	import type { SelectOption, Tag } from '@melt-ui/svelte';
	import { ArrowRight, Copy, Info, MoreHorizontal, Pencil, Trash2, X } from 'lucide-svelte';
	import { writable } from 'svelte/store';

	$: palette = variants[$user.flavour];

	const sections = ['Typography', 'Actions', 'Form', 'Display', 'Overlays', 'Navigation'];
	const tones = Object.values(Tone);
	const sizes = Object.entries(TextSize);
	const weights = Object.entries(TextWeight);
	const colors = Object.entries(TextColor).filter(([name]) => name !== 'Inherit');

	let clicks = 0;
	let lastAction = 'none';
	const text = writable('Hello');
	const email = writable('not-an-email');
	const message = writable('');
	const number = writable(42);
	const checked = writable(true);
	const enabled = writable(true);
	const radio = writable('Medium');
	const framework = writable<SelectOption<string>>({ value: 'svelte', label: 'Svelte' });
	const volume = writable([30]);
	const dialogOpen = writable(false);
	const tab = writable('overview');
	const page = writable(1);

	const tags = writable<Tag[]>([
		{ id: 'svelte', value: 'svelte' },
		{ id: 'go', value: 'go' }
	]);
	const date = writable<DateValue | undefined>(undefined);
	const compactDate = writable<DateValue | undefined>(undefined);
	const color = writable<PaletteColor>('mauve');
	const wheelColor = writable<PaletteColor>('blue');
	let file: File | null = null;

	const frameworks = ['svelte', 'react', 'vue', 'solid', 'angular', 'go', 'gin', 'mongodb'];
	const suggest = async (query: string) =>
		frameworks.filter((name) => name.includes(query.toLowerCase()));

	const menu: MenuItem[] = [
		{ label: 'Edit', icon: Pencil, onSelect: () => (lastAction = 'Edit') },
		{ label: 'Duplicate', icon: Copy, onSelect: () => (lastAction = 'Duplicate') },
		{ label: 'Archive', disabled: true, onSelect: () => (lastAction = 'Archive') },
		{ label: 'Delete', icon: Trash2, tone: Tone.Danger, onSelect: () => (lastAction = 'Delete') }
	];

	const toasts = [
		{ title: 'Saved', description: 'Your changes are stored', color: 'bg-ctp-green' },
		{ title: 'Heads up', description: 'This is only a demo', color: 'bg-ctp-yellow' },
		{ title: 'Failed', description: 'Something went wrong', color: 'bg-ctp-red' }
	];
	const toastTones = [Tone.Success, Tone.Warning, Tone.Danger];
</script>

<svelte:head>
	<title>Playground - Mathéo Galuba</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<Hero
	class="mb-16"
	highlight="Playground"
	description="Every shared component, grouped by purpose, with live state. Only served on localhost."
>
	<nav class="flex flex-wrap justify-center gap-x-6 gap-y-2" aria-label="Sections">
		{#each sections as section}
			<Link href="#{section.toLowerCase()}" weight={TextWeight.Semibold} color={TextColor.Accent}>
				{section}
			</Link>
		{/each}
	</nav>
</Hero>

<Section title="Typography" description="Text is sized, weighted and coloured through enums.">
	<Demo name="Heading">
		{#each [1, 2, 3, 4, 5, 6] as const as level}
			<Heading {level}>Heading {level}</Heading>
		{/each}
		<Heading level={2} gradient>Gradient</Heading>
		<Heading level={3} color={TextColor.Heading}>Article heading</Heading>
	</Demo>

	<Demo name="Text">
		<div class="flex flex-col gap-1">
			{#each sizes as [name, size]}
				<Text {size}>TextSize.{name}</Text>
			{/each}
		</div>
		<div class="flex flex-col gap-1">
			{#each weights as [name, weight]}
				<Text {weight}>TextWeight.{name}</Text>
			{/each}
		</div>
	</Demo>

	<Demo name="Text">
		<div class="flex flex-col gap-1">
			{#each colors as [name, color]}
				<Text
					{color}
					weight={TextWeight.Semibold}
					class={color === TextColor.Inverse ? 'w-fit rounded-md bg-ctp-text px-2' : ''}
				>
					TextColor.{name}
				</Text>
			{/each}
		</div>
	</Demo>

	<Demo name="Link">
		<Text>An <Link href="/">internal link</Link> inside a sentence.</Text>
		<Text>
			An <Link href="https://svelte.dev" external>external link</Link> opens in a new tab.
		</Text>
	</Demo>

	<Demo name="Code">
		<Text>Run <Code>bun run dev</Code> to start the dev server.</Text>
	</Demo>
</Section>

<Section title="Actions">
	<Demo name="Button" wide>
		{#each Object.values(ButtonVariant) as variant}
			<div class="flex flex-wrap items-center gap-3">
				{#each tones as tone}
					<Button {variant} {tone} on:click={() => clicks++}>{tone}</Button>
				{/each}
			</div>
		{/each}
		<div class="flex flex-wrap items-center gap-3">
			<Button size={TextSize.Sm} on:click={() => clicks++}>Small</Button>
			<Button on:click={() => clicks++}>With icon <ArrowRight size="16" /></Button>
			<Button size={TextSize.Lg} on:click={() => clicks++}>Large</Button>
			<Button disabled on:click={() => clicks++}>Disabled</Button>
			<Button loading on:click={() => clicks++}>Loading</Button>
			<Button href="/">Link to home</Button>
		</div>
		<svelte:fragment slot="state">clicks: {clicks}</svelte:fragment>
	</Demo>

	<Demo name="IconButton">
		<div class="flex flex-wrap items-center gap-3">
			<IconButton label="Close" on:click={() => clicks++}><X size="16" /></IconButton>
			<IconButton label="Edit" tone={Tone.Accent} on:click={() => clicks++}>
				<Pencil size="16" />
			</IconButton>
			<IconButton label="Delete" tone={Tone.Danger} on:click={() => clicks++}>
				<Trash2 size="16" />
			</IconButton>
			<IconButton label="Disabled" disabled><X size="16" /></IconButton>
		</div>
	</Demo>

	<Demo name="DropdownMenu">
		<DropdownMenu items={menu} label="Actions" placement="bottom-start">
			<span class="grid place-items-center rounded-md p-1 hover:bg-ctp-surface0">
				<MoreHorizontal size="18" />
			</span>
		</DropdownMenu>
		<svelte:fragment slot="state">last action: {lastAction}</svelte:fragment>
	</Demo>
</Section>

<Section title="Form">
	<Demo name="TextInput">
		<TextInput label="Label" placeholder="Type something" value={text} maxlength={30} />
		<TextInput
			label="Email"
			type="email"
			value={email}
			error={/\S+@\S+\.\S+/.test($email) ? '' : 'Email address is invalid'}
		/>
		<TextInput label="With a hint" hint="Shown under the field" placeholder="Optional" />
		<TextInput label="Disabled" value={text} disabled />
		<svelte:fragment slot="state">value: "{$text}" · email: "{$email}"</svelte:fragment>
	</Demo>

	<Demo name="Textarea">
		<Textarea
			label="Message"
			placeholder="Write a few lines"
			value={message}
			maxlength={200}
			hint="{$message.length} / 200"
		/>
		<Textarea label="Disabled" rows={2} value={writable('Read only')} disabled />
	</Demo>

	<Demo name="NumberInput">
		<NumberInput label="Between 0 and 100, by 5" value={number} min={0} max={100} step={5} />
		<NumberInput label="Disabled" value={number} disabled />
		<svelte:fragment slot="state">value: {$number}</svelte:fragment>
	</Demo>

	<Demo name="Select">
		<Select
			label="Framework"
			options={[
				{ value: 'svelte', label: 'Svelte' },
				{ value: 'react', label: 'React' },
				{ value: 'vue', label: 'Vue' }
			]}
			value={framework}
		/>
		<Select label="Plain strings, label hidden" hideLabel options={['One', 'Two', 'Three']} />
		<Select label="Disabled" options={['One']} disabled />
		<svelte:fragment slot="state">value: {JSON.stringify($framework)}</svelte:fragment>
	</Demo>

	<Demo name="Checkbox">
		<Checkbox name="playground-checkbox" label="I agree" {checked} />
		<Checkbox
			name="playground-checkbox-indeterminate"
			label="Indeterminate by default"
			defaultChecked="indeterminate"
		/>
		<Checkbox name="playground-checkbox-disabled" label="Disabled" disabled />
		<svelte:fragment slot="state">checked: {$checked}</svelte:fragment>
	</Demo>

	<Demo name="Switch">
		<Switch name="playground-switch" label="Notifications" checked={enabled} />
		<Switch name="playground-switch-disabled" label="Disabled" disabled />
		<svelte:fragment slot="state">checked: {$enabled}</svelte:fragment>
	</Demo>

	<Demo name="RadioGroup">
		<RadioGroup name="playground-radio" options={['Small', 'Medium', 'Large']} value={radio} />
		<RadioGroup
			name="playground-radio-vertical"
			options={['Small', 'Medium', 'Large']}
			value={radio}
			orientation="vertical"
		/>
		<svelte:fragment slot="state">value: "{$radio}"</svelte:fragment>
	</Demo>

	<Demo name="Slider">
		<Slider label="Volume" value={volume} format={(value) => `${value}%`} />
		<Slider
			label="With ticks"
			min={0}
			max={4}
			step={1}
			defaultValue={2}
			ticks={['XS', 'S', 'M', 'L', 'XL']}
		/>
		<Slider label="Disabled" disabled />
		<svelte:fragment slot="state">value: {JSON.stringify($volume)}</svelte:fragment>
	</Demo>

	<Demo name="MeltTags" legacy>
		<MeltTags {tags} placeholder="Add a tag, suggestions as you type" {suggest} />
		<MeltTags placeholder="Without suggestions" />
		<svelte:fragment slot="state">
			tags: {JSON.stringify($tags.map((tag) => tag.value))}
		</svelte:fragment>
	</Demo>

	<Demo name="MeltDatePicker" legacy>
		<MeltDatePicker label="Due date" value={date} />
		<div class="flex items-center gap-2">
			<Text as="span" size={TextSize.Sm}>Compact:</Text>
			<MeltDatePicker compact value={compactDate} />
		</div>
		<svelte:fragment slot="state">
			value: {$date?.toString() ?? 'undefined'} · compact: {$compactDate?.toString() ?? 'undefined'}
		</svelte:fragment>
	</Demo>

	<Demo name="FileUploader" legacy>
		<FileUploader
			label="An image"
			accept=".png,.jpg,.jpeg,.svg"
			onFileSelect={(selected) => (file = selected)}
		/>
		<svelte:fragment slot="state">
			file: {file ? `${file.name} (${file.size} bytes)` : 'null'}
		</svelte:fragment>
	</Demo>

	<Demo name="ColorPicker" legacy>
		<ColorPicker label="Background" value={color} {palette} allowTransparent />
		<svelte:fragment slot="state">value: "{$color}"</svelte:fragment>
	</Demo>

	<Demo name="PaletteWheel" legacy>
		<div class="flex justify-center">
			<PaletteWheel label="Accent" value={wheelColor} {palette} />
		</div>
		<svelte:fragment slot="state">value: "{$wheelColor}"</svelte:fragment>
	</Demo>
</Section>

<Section title="Display">
	<Demo name="Card">
		<Card>
			<Heading level={4}>Plain card</Heading>
			<Text color={TextColor.Muted}>The surface used for grouped content.</Text>
		</Card>
		<Card interactive>
			<Heading level={4}>Interactive card</Heading>
			<Text color={TextColor.Muted}>Shows the accent border on hover.</Text>
		</Card>
	</Demo>

	<Demo name="Badge">
		<div class="flex flex-wrap items-center gap-2">
			{#each tones as tone}
				<Badge {tone}>{tone}</Badge>
			{/each}
		</div>
		<div class="flex flex-wrap items-center gap-2">
			<Badge tone={Tone.Success}><Info size="12" /> With icon</Badge>
			<Badge size={TextSize.Sm}>Larger</Badge>
		</div>
	</Demo>

	<Demo name="Progress">
		<Progress label="Upload" value={$volume[0]} />
		<Progress label="Steps" value={2} max={5} />
		<Text size={TextSize.Xs} color={TextColor.Muted}>The first bar follows the Volume slider.</Text>
	</Demo>

	<Demo name="Loader">
		<div class="flex items-end gap-6">
			<Loader size={16} />
			<Loader />
			<Loader size={64} color={TextColor.Muted} />
		</div>
	</Demo>

	<Demo name="Divider">
		<Text>Above</Text>
		<Divider />
		<div class="flex items-center gap-3">
			<Text as="span">Left</Text>
			<Divider vertical />
			<Text as="span">Right</Text>
		</div>
	</Demo>

	<Demo name="ProviderIcon" legacy>
		<div class="flex flex-wrap items-center gap-6">
			{#each ['email', 'github', 'google'] as provider}
				<span class="flex items-center gap-2">
					<ProviderIcon {provider} size={24} />
					<Text as="span">{provider}</Text>
				</span>
			{/each}
		</div>
	</Demo>

	<Demo name="GitHub" legacy>
		<div class="flex items-end gap-6">
			<GitHub size={16} />
			<GitHub />
			<GitHub size={48} strokeWidth={1} />
			<span class={TextColor.Accent}><GitHub size={48} /></span>
		</div>
	</Demo>

	<Demo name="Google" legacy>
		<div class="flex items-end gap-6">
			<Google size={16} />
			<Google />
			<Google size={48} strokeWidth={1} />
			<span class={TextColor.Accent}><Google size={48} /></span>
		</div>
	</Demo>

	<Demo name="CodeBlock" legacy wide>
		<CodeBlock>
			<pre data-lang="go"><code
					>{`func main() {
	fmt.Println("Hello from a code block")
}`}</code
				></pre>
		</CodeBlock>
	</Demo>

	<Demo name="MultiLangCodeBlock" legacy wide>
		<MultiLangCodeBlock>
			<pre data-lang="py"><code>{`print("Hello from Python")`}</code></pre>
			<pre data-lang="js"><code>{`console.log('Hello from JavaScript');`}</code></pre>
			<pre data-lang="rs"><code
					>{`fn main() {
    println!("Hello from Rust");
}`}</code
				></pre>
		</MultiLangCodeBlock>
	</Demo>
</Section>

<Section title="Overlays">
	<Demo name="Dialog">
		<div>
			<Button on:click={() => dialogOpen.set(true)}>Open dialog</Button>
		</div>
		<Dialog open={dialogOpen} title="Delete this item?" description="It will be deleted for good.">
			<svelte:fragment slot="footer">
				<Button
					variant={ButtonVariant.Ghost}
					tone={Tone.Neutral}
					on:click={() => dialogOpen.set(false)}
				>
					Cancel
				</Button>
				<Button tone={Tone.Danger} on:click={() => dialogOpen.set(false)}>Delete</Button>
			</svelte:fragment>
		</Dialog>
		<svelte:fragment slot="state">open: {$dialogOpen}</svelte:fragment>
	</Demo>

	<Demo name="Popover">
		<div>
			<Popover label="Open details">
				<Badge slot="trigger" size={TextSize.Sm}><Info size="14" /> Details</Badge>
				<div class="flex max-w-[14rem] flex-col gap-2">
					<Text weight={TextWeight.Semibold}>Popover</Text>
					<Text size={TextSize.Sm} color={TextColor.Muted}>
						Free content anchored to its trigger.
					</Text>
				</div>
			</Popover>
		</div>
	</Demo>

	<Demo name="Tooltip">
		<div class="flex flex-wrap items-center gap-6">
			{#each ['top', 'bottom', 'left', 'right'] as const as placement}
				<Tooltip text="Plain text" {placement}>
					<Text as="span" weight={TextWeight.Semibold} class="flex items-center gap-1">
						<Info size="16" />
						{placement}
					</Text>
				</Tooltip>
			{/each}
		</div>
		<div class="flex flex-wrap items-center gap-6">
			<Tooltip html="<strong>Bold</strong> and <em>italic</em>">
				<Text as="span" weight={TextWeight.Semibold} class="flex items-center gap-1">
					<Info size="16" /> html
				</Text>
			</Tooltip>
			<Tooltip math={'e^{i\\pi} + 1 = 0'}>
				<Text as="span" weight={TextWeight.Semibold} class="flex items-center gap-1">
					<Info size="16" /> math
				</Text>
			</Tooltip>
		</div>
	</Demo>

	<Demo name="Toast" path="addToast from routes/+layout.svelte">
		<div class="flex flex-wrap items-center gap-3">
			{#each toasts as toast, i}
				<Button tone={toastTones[i]} on:click={() => addToast({ data: toast })}>
					{toast.title}
				</Button>
			{/each}
		</div>
	</Demo>
</Section>

<Section title="Navigation">
	<Demo name="Tabs">
		<Tabs
			tabs={[
				{ id: 'overview', title: 'Overview' },
				{ id: 'details', title: 'Details' },
				{ id: 'history', title: 'History' }
			]}
			value={tab}
			let:id
		>
			<Text>Content of the “{id}” tab.</Text>
		</Tabs>
		<svelte:fragment slot="state">value: "{$tab}"</svelte:fragment>
	</Demo>

	<Demo name="Accordion">
		<Accordion
			items={[
				{ id: 'profile', title: 'Profile' },
				{ id: 'account', title: 'Account' },
				{ id: 'appearance', title: 'Appearance' }
			]}
			expanded={['profile']}
			let:id
		>
			<Text color={TextColor.Muted}>Content of the “{id}” panel.</Text>
		</Accordion>
	</Demo>

	<Demo name="Pagination">
		<Pagination {page} count={95} perPage={10} />
		<svelte:fragment slot="state">page: {$page}</svelte:fragment>
	</Demo>
</Section>
