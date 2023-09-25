<script>
	import '../app.postcss';
	import { createToolbar, melt } from '@melt-ui/svelte';
	import LoginDialog from './LoginDialog.svelte';

	const {
    elements: { root, button, link },
  } = createToolbar();
</script>

<div class="app">
	<nav
		class="min-w-max m-3 flex items-center gap-4 rounded-md bg-ctp-surface0 p-3 text-ctp-text"
		use:melt={$root}
	>
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
  	<div class="flex items-center gap-1" >
			<a href="/" class="" use:melt={$link}>Link1</a>
			<a href="/" class="" use:melt={$link}>Link2</a>
			<a href="/" class="" use:melt={$link}>Link3</a>
			<a href="/" class="" use:melt={$link}>Link4</a>
		</div>
		<div class="ml-auto flex items-center gap-1" >
			<LoginDialog />
		</div>
	</nav>
	<main>
		<slot />
	</main>
</div>
