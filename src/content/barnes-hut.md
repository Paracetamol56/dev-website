---
title: Barnes-Hut
description: Optimization algorithm for n-body simulations
tags:
  - computer-science
  - simulation
  - astronomy
  - physics
listed: true
---

<script>
  import BarnesHut from './barnes-hut/BarnesHut.svelte';
  import Compexity from './barnes-hut/Complexity.svelte';
  import Tree from './barnes-hut/Tree.svelte';
  import Button from '$lib/components/Button.svelte';
</script>

<BarnesHut />

## The problem

The [n-body problem](https://en.wikipedia.org/wiki/N-body_problem) is a classical problem in physics and astronomy. It consists of predicting the motion of a group of objects that interact with each other gravitationally. The problem is to find the positions and velocities of the objects at a given time in the future or the past. The problem has no general analytical solution, so it is usually solved numerically.

Computing the gravitational force between 2 objects is a simple problem and not very computationally expensive. In Newtonian mechanics, the force between 2 objects is given by:

$$
\vec{F} = G \frac{m_1 m_2}{r^2} \hat{r}
$$

Easy, right?

But simulating the motion of a large number of objects , let's say 1000000 stars in a galaxy (witch is relatively small), is a much more difficult problem. The naive approach would be to compute the force between each pair of objects, but we would have to compute all the pairs, so 100000 force calculations **per object**, so $1000000^2 = 10^{12}$ force calculations in total.
Hopefully, there are some optimizations we can do to reduce the number of force calculations.

## Complexity

The green line is the linear complexity, <span class="text-ctp-green">$O(n)$</span> just for reference. The naive approach has a complexity of <span class="text-ctp-red">$O(n^2)$</span>, here in red.

The complexity of the Barnes-Hut algorithm is <span class="text-ctp-blue">$O(n \log n)$</span>, here in blue. We can explain this complexity by looking at the tree structure. The tree has a depth of $\log n$, so in the worst case, we have to explore all the nodes in the tree. And we do that $n$ times, once for each object. So the complexity is <span class="text-ctp-blue">$O(n \log n)$</span>.

<Compexity />

We have to keep in mind that this method for decreasing the complexity comes at a cost. The Barnes-Hut algorithm is an **approximation**. Even if the error is negligible, the most accurate method still is the naive approach.

## Barnes-Hut algorithm

The Barnes-Hut algorithm is an algorithm for n-body simulations. It is based on the idea of grouping objects that are far away from the object we are computing the force on, and approximating the force of the group as a whole. This is done recursively, so we can group groups of groups of groups of objects, and so on.

### The power of trees

The algorithm is based on [quadtrees](https://en.wikipedia.org/wiki/Quadtree) for 2D simulations or [octrees](https://en.wikipedia.org/wiki/Octree) for 3D simulations. These are data structures that partition space into smaller regions. Each region can contain a maximum number of objects, and if that number is exceeded, the region is subdivided into 4 (or 8) smaller regions. This is done recursively until the maximum number of objects per region is not exceeded.

In this example, you can add objects to the simulation by clicking on the canvas.

<Tree />

## Barnes-Hut in action (and in 3D)

This is a simulation of a galaxy with 10000 stars. Their positions and velocities are computed on the CPU, and the rendering is done on the GPU with WebGL.
This component is made in C++ and compiled to WebAssembly.

**Your browser must support WebAssembly. Mobile devices are strongly discouraged.**
**Refresh the page if the simulation is not working on initial load.**

In 3d, the core idea is the same but we use an octree (8 children per node) instead of a quadtree (4 children per node).

<iframe src="/files/barnes-hut/barnes-hut.html" title="Barnes-Hut WASM simulation" style="width: 800px; height: 800px; border: none; margin: 0 auto;"></iframe>

<div class="mt-4 flex justify-center">
  <a
    target="_blank"
    href="https://paracetamol56.itch.io/barnes-hut"
    class="content-ignore flex items-center gap-1 rounded-md bg-ctp-mauve px-3 py-1
      font-semibold text-ctp-mantle
      shadow-md shadow-ctp-crust transition-opacity
      hover:opacity-80 active:opacity-60"
  >
    <span>Run on Itch.io</span>
    <svg class="square-4 fill-current" role="img" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><title>Itch.io</title><path d="M3.13 1.338C2.08 1.96.02 4.328 0 4.95v1.03c0 1.303 1.22 2.45 2.325 2.45 1.33 0 2.436-1.102 2.436-2.41 0 1.308 1.07 2.41 2.4 2.41 1.328 0 2.362-1.102 2.362-2.41 0 1.308 1.137 2.41 2.466 2.41h.024c1.33 0 2.466-1.102 2.466-2.41 0 1.308 1.034 2.41 2.363 2.41 1.33 0 2.4-1.102 2.4-2.41 0 1.308 1.106 2.41 2.435 2.41C22.78 8.43 24 7.282 24 5.98V4.95c-.02-.62-2.082-2.99-3.13-3.612-3.253-.114-5.508-.134-8.87-.133-3.362 0-7.945.053-8.87.133zm6.376 6.477a2.74 2.74 0 0 1-.468.602c-.5.49-1.19.795-1.947.795a2.786 2.786 0 0 1-1.95-.795c-.182-.178-.32-.37-.446-.59-.127.222-.303.412-.486.59a2.788 2.788 0 0 1-1.95.795c-.092 0-.187-.025-.264-.052-.107 1.113-.152 2.176-.168 2.95v.005l-.006 1.167c.02 2.334-.23 7.564 1.03 8.85 1.952.454 5.545.662 9.15.663 3.605 0 7.198-.21 9.15-.664 1.26-1.284 1.01-6.514 1.03-8.848l-.006-1.167v-.004c-.016-.775-.06-1.838-.168-2.95-.077.026-.172.052-.263.052a2.788 2.788 0 0 1-1.95-.795c-.184-.178-.36-.368-.486-.59-.127.22-.265.412-.447.59a2.786 2.786 0 0 1-1.95.794c-.76 0-1.446-.303-1.948-.793a2.74 2.74 0 0 1-.468-.602 2.738 2.738 0 0 1-.463.602 2.787 2.787 0 0 1-1.95.794h-.16a2.787 2.787 0 0 1-1.95-.793 2.738 2.738 0 0 1-.464-.602zm-2.004 2.59v.002c.795.002 1.5 0 2.373.953.687-.072 1.406-.108 2.125-.107.72 0 1.438.035 2.125.107.873-.953 1.578-.95 2.372-.953.376 0 1.876 0 2.92 2.934l1.123 4.028c.832 2.995-.266 3.068-1.636 3.07-2.03-.075-3.156-1.55-3.156-3.025-1.124.184-2.436.276-3.748.277-1.312 0-2.624-.093-3.748-.277 0 1.475-1.125 2.95-3.156 3.026-1.37-.004-2.468-.077-1.636-3.072l1.122-4.027c1.045-2.934 2.545-2.934 2.92-2.934zM12 12.714c-.002.002-2.14 1.964-2.523 2.662l1.4-.056v1.22c0 .056.56.033 1.123.007.562.026 1.124.05 1.124-.008v-1.22l1.4.055C14.138 14.677 12 12.713 12 12.713z"/></svg>
  </a>
</div>

## Other use cases

Since the space is partitioned in an esay to use data structure, we can use it for other things, like collision detection. or serching for objects by position. And basically any other problem where you want to avoid looping over all the objects and simplify computations by grouping objects that are far away.

## References

- Original paper: [A Hierarchical O(N log N) Force Calculation Algorithm](https://ui.adsabs.harvard.edu/abs/1986Natur.324..446B/abstract)
- [Wikipedia](https://en.wikipedia.org/wiki/Barnes%E2%80%93Hut_simulation)
- Barnes hut in action: [The Barnes-Hut Approximation](https://jheer.github.io/barnes-hut) by Jeff Heer
- 3D online simulation: [Barnes-Hut Simulation](https://jurasic.dev/2023/barnes-hut-simulation/) by Jurasic Dev
