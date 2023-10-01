<script lang="ts">
	import { createToolbar, melt } from "@melt-ui/svelte";
	import ThemePopover from "./ThemePopover.svelte";
	import type { Writable } from "svelte/store";
	import LoginDialog from "./LoginDialog.svelte";
	import { user } from "$lib/stores";
	import ProfilePopover from "./ProfilePopover.svelte";
	import { ExternalLink } from "lucide-svelte";

  export let theme: Writable<string>;

  const {
    elements: { root, link },
  } = createToolbar();
</script>

<nav
	class="min-w-max p-3"
	use:melt={$root}
>
	<div class="flex items-center gap-4 rounded-md bg-ctp-mantle p-3 text-ctp-text">
		<a href="/">
			<svg class="square-8" width="750" height="606" viewBox="0 0 750 606" fill="none" xmlns="http://www.w3.org/2000/svg">
				<g clip-path="url(#clip0_22_3)">
					<path class="fill-ctp-text" d="M240.326 193.691L355.66 0H750V121.2H427.826L312.5 314.891L240.326 193.691Z"/>
					<path class="fill-ctp-text" d="M312.5 436.024L125 121.2H0L125 363.6L240.326 557.284L312.5 436.024Z"/>
					<path class="fill-ctp-text" d="M750 181.733V606H437.5L500 484.8H625V302.933L536.049 303L608.236 181.766L750 181.733Z"/>
					<path class="fill-ctp-subtext0" d="M463.875 181.8L240.326 557.284L269.34 606H355.625L608.236 181.766L463.875 181.8Z"/>
					<path class="fill-ctp-subtext0" d="M0 606H125V363.6L0 121.2V606Z"/>
				</g>
				<defs>
					<clipPath id="clip0_22_3">
						<rect width="750" height="606" fill="white"/>
					</clipPath>
				</defs>
			</svg>
		</a>
		<div class="flex items-center gap-3 text-md font-semibold" >
			<a class="hover:text-ctp-blue transition-colors" href="/map" use:melt={$link}>Main map</a>
			<a class="hover:text-ctp-blue transition-colors" href="/contact" use:melt={$link}>Contact</a>
			<a class="hover:text-ctp-blue transition-colors hidden sm:inline-flex items-center gap-1" href="https://matheo-galuba.com" target="_blank" use:melt={$link}>Portfolio <ExternalLink size="16" stroke-width="3" /></a>
		</div>
		<div class="ml-auto flex items-center gap-2" >
			<ThemePopover {theme} />
			{#if $user !== null}
			<ProfilePopover />
			{:else}
			<LoginDialog />
			{/if}
		</div>
	</div>
</nav>
