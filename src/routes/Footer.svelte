<script lang="ts">
	import SocialButton from './SocialButton.svelte';
</script>

<footer class="min-w-max p-3">
	<div class="flex flex-col items-center gap-4 rounded-md bg-ctp-mantle p-3 text-ctp-text">
		<p class="font-semibold">
      Source code available on <a class="text-ctp-blue" href="https://github.com/Paracetamol56/dev-showcase">GitHub</a>.
    </p>
		<div class="flex gap-2">
			<SocialButton href="https://github.com/Paracetamol56" tooltip="GitHub">
				<svg
					class="square-8"
					clip-rule="evenodd"
					fill-rule="evenodd"
					stroke-linejoin="round"
					stroke-miterlimit="2"
					viewBox="0 0 64 64"
					xmlns="http://www.w3.org/2000/svg"
				>
					<ellipse
						cx="21.624"
						cy="22.613"
						fill="currentColor"
						rx="11.099"
						ry="8.306"
						transform="matrix(2.88315 0 0 3.85264 -30.3444 -55.1209)"
					/>
					<path
						class="fill-ctp-mantle"
						d="m13.257 21.063c.058-.362.225-.995.712-1.481 0 0-5.41-.027-5.41-5.738 0-2.01 1.19-3.18 1.19-3.18s-.495-1.777.135-3.108c0 0 1.268-.344 3.285 1.221 0 0 1.367-.471 2.951-.437 1.585-.034 2.951.437 2.951.437 2.018-1.565 3.286-1.221 3.286-1.221.629 1.331.135 3.108.135 3.108s1.19 1.17 1.19 3.18c0 5.711-5.41 5.738-5.41 5.738.722.722.741 1.767.741 1.767v3.772c0 .381-.39.391-.39.391h-5.005s-.391-.01-.391-.391v-2.23c-.322.098-.789.207-1.275.207-2.767 0-2.69-2.213-3.471-2.854-.525-.43-1.138-.889-1.138-1.048 0-.079-.038-.264.51-.264 1.854 0 1.737 2.515 4.023 2.515.445 0 1.063-.244 1.381-.384z"
						transform="matrix(2.07452 0 0 2.00422 -.180363 -.105643)"
					/>
				</svg>
			</SocialButton>
			<SocialButton href="https://www.linkedin.com/in/matheogaluba/" tooltip="LinkedIn">
        <p>LinkedIn</p>
      </SocialButton>
			<SocialButton href="mailto:matheo.galu56@gmail.com" tooltip="Email">
				<p>Email</p>
			</SocialButton>
			<SocialButton href="https://t.me/paracetamol_56" tooltip="Telegram">
				<p>Telegram</p>
			</SocialButton>
			<SocialButton href="https://discordapp.com/users/609748578004500517/" tooltip="Discord">
				<svg
          class="square-8"
					clip-rule="evenodd"
					fill-rule="evenodd"
					stroke-linejoin="round"
					stroke-miterlimit="2"
					viewBox="0 0 64 64"
					xmlns="http://www.w3.org/2000/svg"
					><ellipse
						cx="21.624"
						cy="22.613"
						fill="currentColor"
						rx="11.099"
						ry="8.306"
						transform="matrix(2.88315 0 0 3.85264 -30.3444 -55.1209)"
					/><path
            class="fill-ctp-mantle"
						d="m55.738 4.498c-.013-.024-.034-.044-.06-.054-2.978-1.367-6.121-2.341-9.35-2.899-.06-.012-.121.017-.15.071-.428.777-.817 1.574-1.164 2.39-3.481-.528-7.022-.528-10.502 0-.35-.818-.745-1.616-1.183-2.39-.031-.053-.09-.081-.15-.071-3.23.556-6.373 1.531-9.351 2.899-.025.01-.047.029-.061.052-5.956 8.894-7.587 17.569-6.787 26.136.005.043.026.082.06.108 3.468 2.568 7.346 4.527 11.471 5.796.059.018.123-.003.161-.053.886-1.205 1.671-2.481 2.346-3.816.04-.077.003-.172-.079-.202-1.238-.473-2.436-1.044-3.584-1.707-.09-.053-.097-.181-.014-.244.241-.181.482-.368.712-.558.041-.034.099-.042.148-.02 7.518 3.431 15.657 3.431 23.086 0 .049-.023.108-.016.15.018.23.19.471.379.714.56.084.062.077.19-.013.244-1.145.668-2.344 1.239-3.585 1.705-.082.031-.118.127-.078.204.687 1.327 1.471 2.602 2.345 3.814.036.051.101.073.161.055 4.131-1.265 8.017-3.225 11.489-5.796.034-.025.056-.064.06-.106.958-9.905-1.604-18.508-6.792-26.136zm-23.597 20.918c-2.263 0-4.128-2.077-4.128-4.628s1.829-4.628 4.128-4.628c2.318 0 4.165 2.095 4.129 4.628 0 2.551-1.829 4.628-4.129 4.628zm15.264 0c-2.263 0-4.128-2.077-4.128-4.628s1.829-4.628 4.128-4.628c2.318 0 4.165 2.095 4.129 4.628 0 2.551-1.811 4.628-4.129 4.628z"
						fill-rule="nonzero"
						transform="matrix(.891251 0 0 .891251 -3.4353 14.0286)"
					/></svg
				>
			</SocialButton>
		</div>
    <small>
      © 2023 Matheo Galuba. All rights reserved.
    </small>
	</div>
</footer>
