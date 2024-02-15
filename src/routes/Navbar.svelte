<script lang="ts">
	import { createToolbar, melt } from '@melt-ui/svelte';
	import ThemePopover from './ThemePopover.svelte';
	import LoginDialog from './LoginDialog.svelte';
	import { user } from '$lib/store';
	import ProfilePopover from './ProfilePopover.svelte';
	import { ExternalLink } from 'lucide-svelte';

	const {
		elements: { root, link }
	} = createToolbar();

	let scrollY: number = 0;
	let scrollYdelta: number = 0;
	const handleMouseWheel = (e: WheelEvent) => {
		scrollYdelta = e.deltaY;
	};
</script>

<svelte:window on:wheel={handleMouseWheel} bind:scrollY />

<nav
	class="min-w-max p-3 fixed top-0 left-0 right-0 transition-transform duration-200
				transform {scrollYdelta > 0 && scrollY >= 250 ? '-translate-y-full' : ''} z-20"
	use:melt={$root}
>
	<div
		class="flex items-center gap-4 rounded-md p-3 text-ctp-text
					shadow-md shadow-ctp-crust backdrop-blur-lg bg-ctp-crust/50"
	>
		<a href="/">
			<svg
				class="square-8"
				version="1.1"
				xmlns="http://www.w3.org/2000/svg"
				xmlns:xlink="http://www.w3.org/1999/xlink"
				viewBox="0 0 769.3 606"
				xml:space="preserve"
			>
				<g>
					<defs>
						<rect id="SVGID_1_" width="769.3" height="606" />
					</defs>
					<clipPath id="SVGID_00000043451981554379318190000013841271621363195293_">
						<use xlink:href="#SVGID_1_" style="overflow:visible;" />
					</clipPath>
					<g style="clip-path:url(#SVGID_00000043451981554379318190000013841271621363195293_);">
						<path class="fill-ctp-text" d="M259.7,193.7L375,0h394.3v121.2H447.2L331.8,314.9L259.7,193.7z" />
						<path class="fill-ctp-text" d="M331.8,436L144.3,121.2h-125H0l259.7,436.1L331.8,436z" />
						<path
							class="fill-ctp-text"
							d="M769.3,181.7V606H456.8l62.5-121.2h125V302.9l-89,0.1l72.2-121.2L769.3,181.7z"
						/>
						<path class="fill-ctp-subtext0" d="M483.2,181.8L259.7,557.3l29,48.7H375l252.6-424.2L483.2,181.8z" />
						<path class="fill-ctp-subtext0" d="M0,606h125V331.1L0,121.2V606z" />
					</g>
				</g>
			</svg>
		</a>
		<div class="flex items-center gap-3 text-md font-semibold">
			<a class="hover:text-ctp-blue transition-colors" href="/contact" use:melt={$link}>Contact</a>
			<a
				class="hover:text-ctp-blue transition-colors flex items-center gap-1"
				href="https://matheo-galuba.com"
				target="_blank"
				use:melt={$link}>Portfolio <ExternalLink size="16" stroke-width="3" /></a
			>
		</div>
		<div class="ml-auto flex items-center gap-2">
			<ThemePopover />
			{#if $user.id !== null}
				<ProfilePopover />
			{:else}
				<LoginDialog />
			{/if}
		</div>
	</div>
</nav>
