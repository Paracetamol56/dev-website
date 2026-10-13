<script lang="ts">
	import { socials } from '$lib/socials';
	import SocialButton from './SocialButton.svelte';
</script>

<footer class="sticky top-[100vh] p-3">
	<div
		class="flex flex-col items-center gap-4 rounded-md bg-ctp-mantle p-3 text-center text-ctp-text shadow-md shadow-ctp-crust"
	>
		<p class="font-semibold">
			Source code available on <a
				class="text-ctp-blue"
				href="https://github.com/Paracetamol56/dev-showcase"
				target="_blank"
			>
				GitHub
			</a>
		</p>
		<div class="flex flex-wrap justify-center gap-2">
			{#each socials as social}
				<SocialButton href={social.href} tooltip={social.name}>
					<svg xmlns="http://www.w3.org/2000/svg" class="square-8" viewBox="0 0 64 64">
						<circle class="fill-current" cx="32" cy="32" r="32" />
						{#each social.paths as d}
							<path class="fill-ctp-mantle" {d} />
						{/each}
					</svg>
				</SocialButton>
			{/each}
		</div>
		<small>
			© {new Date().getFullYear()} Matheo Galuba. All rights reserved.
			<a class="text-ctp-blue" href="/pages/privacy-policy">Privacy policy</a></small
		>
	</div>
</footer>
