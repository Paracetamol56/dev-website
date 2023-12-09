<script lang="ts">
	import SocialButton from './SocialButton.svelte';
</script>

<footer class="sticky top-[100vh] min-w-max p-3">
	<div
		class="flex flex-col items-center gap-4 rounded-md bg-ctp-mantle p-3 text-ctp-text shadow-md shadow-ctp-crust"
	>
		<p class="font-semibold">
			Source code available on <a
				class="text-ctp-blue"
				href="https://github.com/Paracetamol56/dev-showcase"
				target="_blank"
			>
				GitHub
			</a>
			.
		</p>
		<div class="flex gap-2">
			<SocialButton href="https://github.com/Paracetamol56" tooltip="GitHub">
				<svg
					version="1.1"
					xmlns="http://www.w3.org/2000/svg"
					xmlns:xlink="http://www.w3.org/1999/xlink"
					x="0px"
					y="0px"
					class="square-8"
					viewBox="0 0 64 64"
					style="enable-background:new 0 0 64 64;"
					xml:space="preserve"
				>
					<circle class="fill-current" cx="32" cy="32" r="32" />
					<path
						class="fill-ctp-mantle"
						d="M25.7,50.7c0-1.2,0-3.4,0-4.9c-6.4,1.2-8.1-1.6-8.6-3c-0.3-0.7-1.5-3-2.6-3.6c-0.9-0.5-2.2-1.7,0-1.7
					c2,0,3.4,1.9,3.9,2.6c2.3,3.9,6,2.8,7.4,2.1c0.2-1.7,0.9-2.8,1.6-3.4c-5.7-0.6-11.6-2.8-11.6-12.6c0-2.8,1-5.1,2.6-6.9
					c-0.3-0.6-1.1-3.3,0.3-6.8c0,0,2.1-0.7,7,2.6c2-0.6,4.2-0.9,6.4-0.9c2.2,0,4.3,0.3,6.4,0.9c4.9-3.3,7-2.6,7-2.6
					c1.4,3.5,0.5,6.1,0.3,6.8c1.6,1.8,2.6,4.1,2.6,6.9c0,9.8-6,12-11.7,12.6c0.9,0.8,1.7,2.3,1.7,4.7c0,2.6,0,5.8,0,7.2
					c0,0.4-0.3,0.8-0.8,0.8H26.5C26.1,51.5,25.7,51.2,25.7,50.7z"
					/>
				</svg>
			</SocialButton>
			<SocialButton href="https://www.linkedin.com/in/matheogaluba/" tooltip="LinkedIn">
				<svg
					version="1.1"
					xmlns="http://www.w3.org/2000/svg"
					xmlns:xlink="http://www.w3.org/1999/xlink"
					x="0px"
					y="0px"
					class="square-8"
					viewBox="0 0 64 64"
					style="enable-background:new 0 0 64 64;"
					xml:space="preserve"
				>
					<circle class="fill-current" cx="32" cy="32" r="32" />
					<path
						class="fill-ctp-mantle"
						d="M22.4,48.8V26.1h-7.5v22.6H22.4z M18.7,23c2.6,0,4.3-1.7,4.3-3.9c0-2.2-1.6-3.9-4.2-3.9
						c-2.6,0-4.3,1.7-4.3,3.9C14.4,21.3,16.1,23,18.7,23L18.7,23L18.7,23z"
					/>
					<path
						class="fill-ctp-mantle"
						d="M26.6,48.8h7.5V36.1c0-0.7,0-1.4,0.2-1.8c0.5-1.4,1.8-2.8,3.9-2.8c2.7,0,3.8,2.1,3.8,5.1v12.1h7.5
						v-13c0-7-3.7-10.2-8.7-10.2c-4.1,0-5.8,2.3-6.8,3.8h0.1v-3.3h-7.5C26.7,28.3,26.6,48.8,26.6,48.8L26.6,48.8z"
					/>
				</svg>
			</SocialButton>
			<SocialButton href="mailto:matheo.galu56@gmail.com" tooltip="Email">
				<svg
					version="1.1"
					xmlns="http://www.w3.org/2000/svg"
					xmlns:xlink="http://www.w3.org/1999/xlink"
					x="0px"
					y="0px"
					class="square-8"
					viewBox="0 0 64 64"
					style="enable-background:new 0 0 64 64;"
					xml:space="preserve"
				>
					<circle class="fill-current" cx="32" cy="32" r="32" />
					<path
						class="fill-ctp-mantle"
						d="M32,14.4c-9.7,0-17.6,7.9-17.6,17.6S22.3,49.6,32,49.6c3.8,0,7.5-1.2,10.5-3.5c1-0.7,1.2-2.1,0.4-3.1
					c-0.7-1-2.1-1.2-3.1-0.4c-2.3,1.7-5.1,2.6-7.9,2.6c-7.3,0-13.2-5.9-13.2-13.2S24.7,18.8,32,18.8S45.2,24.7,45.2,32v1.5
					c0,1.3-1.1,2.4-2.4,2.4s-2.4-1.1-2.4-2.4v-7.7c0-1.2-1-2.2-2.2-2.2c-0.8,0-1.6,0.5-1.9,1.2c-1.2-0.7-2.7-1.2-4.2-1.2
					c-4.6,0-8.3,3.7-8.3,8.3s3.7,8.3,8.3,8.3c2.2,0,4.2-0.9,5.7-2.3c1.2,1.4,3,2.3,5,2.3c3.7,0,6.8-3,6.8-6.8V32
					C49.6,22.3,41.7,14.4,32,14.4z M32,36c-2.2,0-4-1.8-4-4s1.8-4,4-4s4,1.8,4,4S34.2,36,32,36z"
					/>
				</svg>
			</SocialButton>
			<SocialButton href="https://t.me/paracetamol_56" tooltip="Telegram">
				<svg
					version="1.1"
					xmlns="http://www.w3.org/2000/svg"
					xmlns:xlink="http://www.w3.org/1999/xlink"
					x="0px"
					y="0px"
					class="square-8"
					viewBox="0 0 64 64"
					style="enable-background:new 0 0 64 64;"
					xml:space="preserve"
				>
					<circle class="fill-current" cx="32" cy="32" r="32" />
					<path
						class="fill-ctp-mantle"
						d="M12.7,30c9.4-4.1,15.7-6.8,18.8-8.1c9-3.7,10.8-4.4,12.1-4.4c0.3,0,0.9,0.1,1.3,0.4
					c0.3,0.3,0.4,0.6,0.5,0.9s0.1,0.8,0.1,1.3c-0.5,5.1-2.6,17.5-3.7,23.2c-0.5,2.4-1.3,3.2-2.2,3.3c-1.9,0.2-3.3-1.2-5.1-2.4
					c-2.9-1.9-4.4-3-7.2-4.8c-3.2-2.1-1.1-3.3,0.7-5.1c0.5-0.5,8.7-8,8.9-8.7c0-0.1,0-0.4-0.2-0.6c-0.2-0.2-0.5-0.1-0.7-0.1
					c-0.3,0.1-4.8,3.1-13.6,9c-1.3,0.9-2.4,1.3-3.5,1.3c-1.2,0-3.4-0.6-5-1.2c-2-0.6-3.6-1-3.5-2.1C10.4,31.2,11.2,30.6,12.7,30z"
					/>
				</svg>
			</SocialButton>
			<SocialButton href="https://discordapp.com/users/609748578004500517/" tooltip="Discord">
				<svg
					version="1.1"
					xmlns="http://www.w3.org/2000/svg"
					xmlns:xlink="http://www.w3.org/1999/xlink"
					x="0px"
					y="0px"
					class="square-8"
					viewBox="0 0 64 64"
					style="enable-background:new 0 0 64 64;"
					xml:space="preserve"
				>
					<circle class="fill-current" cx="32" cy="32" r="32" />
					<path
						class="fill-ctp-mantle"
						d="M45.2,19.9c-2.4-1.1-5-1.9-7.8-2.4c0,0-0.1,0-0.1,0.1c-0.3,0.6-0.7,1.4-1,2c-2.9-0.4-5.9-0.4-8.7,0
					c-0.3-0.6-0.6-1.4-1-2c0,0-0.1-0.1-0.1-0.1c-2.7,0.5-5.3,1.3-7.8,2.4c0,0,0,0-0.1,0c-4.9,7.4-6.3,14.6-5.6,21.7c0,0,0,0.1,0,0.1
					c3.3,2.4,6.4,3.9,9.5,4.8c0,0,0.1,0,0.1,0c0.7-1,1.4-2.1,1.9-3.2c0-0.1,0-0.1-0.1-0.2c-1-0.4-2-0.9-3-1.4c-0.1,0-0.1-0.2,0-0.2
					c0.2-0.1,0.4-0.3,0.6-0.5c0,0,0.1,0,0.1,0c6.2,2.9,13,2.9,19.2,0c0,0,0.1,0,0.1,0c0.2,0.2,0.4,0.3,0.6,0.5c0.1,0.1,0.1,0.2,0,0.2
					c-1,0.6-1.9,1-3,1.4c-0.1,0-0.1,0.1-0.1,0.2c0.6,1.1,1.2,2.2,1.9,3.2c0,0,0.1,0.1,0.1,0c3.1-1,6.3-2.4,9.5-4.8c0,0,0-0.1,0-0.1
					C51.7,33.4,49.6,26.3,45.2,19.9C45.3,19.9,45.2,19.9,45.2,19.9z M25.7,37.3c-1.9,0-3.4-1.7-3.4-3.8s1.5-3.8,3.4-3.8
					c1.9,0,3.5,1.7,3.4,3.8C29.1,35.6,27.6,37.3,25.7,37.3z M38.4,37.3c-1.9,0-3.4-1.7-3.4-3.8s1.5-3.8,3.4-3.8c1.9,0,3.5,1.7,3.4,3.8
					C41.8,35.6,40.3,37.3,38.4,37.3z"
					/>
				</svg>
			</SocialButton>
		</div>
		<small> © 2023 Matheo Galuba. All rights reserved. </small>
	</div>
</footer>
