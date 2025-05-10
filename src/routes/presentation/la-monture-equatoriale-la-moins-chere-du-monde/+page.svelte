<script lang="ts">
	import Reveal from 'reveal.js';
	import Highlight from 'reveal.js/plugin/highlight/highlight';
	import Markdown from 'reveal.js/plugin/markdown/markdown';
	import Notes from 'reveal.js/plugin/notes/notes';
	import 'reveal.js/dist/reset.css';
	import 'reveal.js/dist/reveal.css';
	import 'reveal.js/dist/theme/black.css';
	import 'reveal.js/plugin/highlight/monokai.css';

	import { onMount } from 'svelte';

	onMount(() => {
		const deck = new Reveal({
			plugins: [Markdown, Highlight, Notes],
			autoAnimateEasing: 'ease',
			autoAnimateDuration: 1,
			hash: true,
			slideNumber: true
		});

		deck.initialize();
	});
</script>

<div class="reveal">
	<div class="slides">
		<section>
			<h2>
				<span class="r-fit-text">La monture équatoriale</span>
				<br />
				<span class="r-fit-text">la moins chère du monde*</span>
			</h2>
			<p><a href="https://matheo-galuba.com">Mathéo Galuba</a> - 2025-05-10</p>
		</section>
		<section>
			<h3>Présentation</h3>
			<section>
				<div class="flex justify-around">
					<img
						class="h-[450px]"
						data-src="/img/la-monture-equatoriale-la-moins-chere-du-monde/20250509_192917.jpg"
						alt="Présentation de la monture 1"
						data-preload
						data-preview-image
					/>
					<img
						class="h-[450px]"
						data-src="/img/la-monture-equatoriale-la-moins-chere-du-monde/20250510_0009.jpg"
						alt="Présentation de la monture 2"
						data-preload
						data-preview-image
					/>
					<img
						class="h-[450px]"
						data-src="/img/la-monture-equatoriale-la-moins-chere-du-monde/20250510_0010.jpg"
						alt="Présentation de la monture 3"
						data-preload
						data-preview-image
					/>
				</div>
			</section>
			<section>
				<div class="flex justify-around">
					<img
						class="h-[450px] mx-auto"
						data-src="/img/la-monture-equatoriale-la-moins-chere-du-monde/20250510_0011.jpg"
						alt="Présentation de la monture 4"
						data-preload
						data-preview-image
					/>
				</div>
			</section>
		</section>
		<section>
			<h2>Liste du matériel</h2>
			<section>
				<h3>Mécanique</h3>
				<ul>
					<li>2 roulements à billes (10-30-9)</li>
					<li>Tube (10 mm)</li>
					<li>Visserie (M2.5)</li>
					<li>Tige filetée M8</li>
					<li>Des trucs lourds</li>
				</ul>
			</section>
			<section>
				<h3>Électronique</h3>
				<ul>
					<li>Un moteur chinois (28BYJ-48 5V)</li>
					<li>Driver ULN2003</li>
					<li>Arduino Nano <i>ou</i> Pro Mini</li>
					<li>Interrupteur(s)</li>
					<li>LED (optionnelle)</li>
					<li>De quoi alimenter l'appareil en 5V</li>
				</ul>
			</section>
		</section>
		<section>
			<h2>Impression 3D</h2>
			<div class="flex justify-around">
				<img
					class="h-[450px]"
					data-src="/img/la-monture-equatoriale-la-moins-chere-du-monde/20250509_143753.png"
					alt="Modèle 3D de la monture"
					data-preload
					data-preview-image
				/>
			</div>
			<p>Source : <a href="https://cults3d.com/:283396">https://cults3d.com/:283396</a></p>
		</section>
		<section>
			<h3>À votre tour !</h3>
			<p>*Offre soumise à conditions :</p>
			<ul>
				<li>Imprimer les pièces</li>
				<li>Assembler le tout</li>
				<li>
					Connecter les composants électroniques<br /><small>Soudure et tout le bazar</small>
				</li>
				<li>Programmer le microcontrôleur</li>
			</ul>
		</section>
		<section>
			<h4>Crédits</h4>
			<p>Créateur : <a href="https://www.instructables.com/member/SimonRob/">SimonRob</a></p>
			<p>
				Article (avec instructions) :<br /><a
					href="https://www.instructables.com/Equatorial-Mount-for-DSLR/"
					>https://www.instructables.com/Equatorial-Mount-for-DSLR/</a
				>
			</p>
		</section>
		<section>
			<h3>Merci !</h3>
		</section>
	</div>
</div>
