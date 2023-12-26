---
title: Boids
description: A simulation of flocking birds
categories:
  - computer science
  - simulation
published: true
---

<script>
  import Boids from './boids/Boids.svelte';
  import CodeBlock from '$lib/components/CodeBlock.svelte';
  import MultiLangCodeBlock from '$lib/components/MultiLangCodeBlock.svelte';
</script>

## What are boids?

Animals that flock together, such as birds, fish, and insects, exhibit emergent behavior. This means that the behavior of the group as a whole is not explicitly programmed, but rather emerges from the behavior of the individuals in the group. The term "boids" was coined by [Craig Reynolds](https://fr.wikipedia.org/wiki/Craig_Reynolds) in 1986 to describe a computer simulation of flocking birds.
This simulation is used in many video games and movies to create realistic-looking crowds.

### Emergent behavior

Crowd intelligence, artificial intelligence, protein folding, or life on earth are examples of emergent behavior. Emergence is a philosophical concept that describes how complex systems and patterns arise out of a multiplicity of relatively simple interactions. The whole is greater than the sum of its parts.

<Boids/>

Very good video about boids by Mehdi Moussaid from [Fouloscopie](https://www.youtube.com/@Fouloscopie) in 5 levels of difficulty (in French):

<div class="flex justify-center">
  <iframe width="560" height="315" src="https://www.youtube.com/embed/Ch7VxxTBe1c?si=OUbHv3_f3D8n98Wa" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>
</div>

## The rules of boids

A boid's behavior is governed by three simple rules. I will consider theses rules as forces that act on the boid because that's pretty much what they are from a physics point of view.

### Cohesion

The first rule is cohesion. A boid will try to move towards the center of mass of the flock. This is the force that keeps the flock together.

<MultiLangCodeBlock>

```ts
function cohesionRule(boid: Boid, otherBoids: Boid[]): number[] {
	const centerOfMass: number[] = [0, 0];
	let numNeighbors: number = 0;

	for (let i = 0; i < otherBoids.length; i++) {
		if (rayDistance(boid, otherBoids[i]) < VISION) {
			centerOfMass[0] += otherBoids[i].x;
			centerOfMass[1] += otherBoids[i].y;
			numNeighbors++;
		}
	}

	if (numNeighbors === 0) {
		return [0, 0];
	}
	centerOfMass[0] /= numNeighbors;
	centerOfMass[1] /= numNeighbors;

	return [centerOfMass[0] - boid.x, centerOfMass[1] - boid.y];
}
```

</MultiLangCodeBlock>

### Separation

Then there is separation. A boid will try to avoid colliding with its neighbors. This is the force that keeps the flock from clumping together.
Its influence is inversely proportional to the distance between the boid and its neighbors and is capped at a certain distance in order to not be in conflict with the cohesion rule.

<MultiLangCodeBlock>

```ts
function separationRule(boid: Boid, otherBoids: Boid[]): number[] {
	const v: number[] = [0, 0];
	for (let i = 0; i < otherBoids.length; i++) {
		const distance: number = rayDistance(boid, otherBoids[i]);
		if (distance === 0) {
			continue;
		}
		if (distance < SEPARATION_VISION) {
			v[0] += (boid.x - otherBoids[i].x) / distance;
			v[1] += (boid.y - otherBoids[i].y) / distance;
		}
	}
	return v;
}
```

</MultiLangCodeBlock>

### Alignment

Lastly, there is alignment. A boid will try to match the average direction of its neighbors. This is the force that keeps the flock moving in the same direction.

<MultiLangCodeBlock>

```ts
function alignmentRule(boid: Boid, otherBoids: Boid[]): number[] {
	const avgV: number[] = [0, 0];
	let numNeighbors: number = 0;

	for (let i = 0; i < otherBoids.length; i++) {
		if (rayDistance(boid, otherBoids[i]) < VISION) {
			avgV[0] += otherBoids[i].vx;
			avgV[1] += otherBoids[i].vy;
			numNeighbors++;
		}
	}

	if (numNeighbors === 0) {
		return [0, 0];
	}
	avgV[0] /= numNeighbors;
	avgV[1] /= numNeighbors;
	return [avgV[0] - boid.vx, avgV[1] - boid.vy];
}
```

</MultiLangCodeBlock>

## Connecting the dots

The 3 functions above return a vector that represents the force that acts on the boid. The boid's velocity is then updated by adding this vector to it. The boid's position is then updated by adding its velocity to it.

I also added 3 constants to tweak the simulation: `COHESION_WEIGHT`, `SEPARATION_WEIGHT`, and `ALIGNMENT_WEIGHT`. This allows us to control the relative influence of each rule on the boid's behavior.

<MultiLangCodeBlock>

```ts
const COHESION_WEIGHT: number = 0.01;
const SEPARATION_WEIGHT: number = 0.1;
const ALIGNMENT_WEIGHT: number = 0.1;
const SPEED: number = 5;

function updateBoid(boid: Boid, otherBoids: Boid[]): void {
	const cohesion: number[] = cohesionRule(boid, otherBoids);
	const separation: number[] = separationRule(boid, otherBoids);
	const alignment: number[] = alignmentRule(boid, otherBoids);

	boid.vx +=
		COHESION_WEIGHT * cohesion[0] +
		SEPARATION_WEIGHT * separation[0] +
		ALIGNMENT_WEIGHT * alignment[0];
	boid.vy +=
		COHESION_WEIGHT * cohesion[1] +
		SEPARATION_WEIGHT * separation[1] +
		ALIGNMENT_WEIGHT * alignment[1];

	boid.x += boid.vx;
	boid.y += boid.vy;

	// Normalize velocity
	const norm: number = Math.sqrt(boid.vx * boid.vx + boid.vy * boid.vy);
	boid.vx = (boid.vx / norm) * SPEED;
	boid.vy = (boid.vy / norm) * SPEED;

	// Draw the boid
	// [...]
}
```

</MultiLangCodeBlock>
