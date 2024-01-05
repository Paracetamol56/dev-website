---
title: Barnes-Hut
description: Optimization algorithm for n-body simulations
categories:
  - computer science
  - simulation
  - astronomy
  - physics
published: true
---

<script>
  import BarnesHut from './barnes-hut/BarnesHut.svelte';
  import Compexity from './barnes-hut/Complexity.svelte';
  import Tree from './barnes-hut/Tree.svelte';
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

This is a simulation of a galaxy with 1000 stars. Their positions and velocities are computed on the CPU, and the rendering is done on the GPU with WebGL.
This component is made in C++ and compiled to WebAssembly.

**Your browser must support WebAssembly. Mobile devices are strongly discouraged.**

In 3d, the core idea is the same but we use an octree (8 children per node) instead of a quadtree (4 children per node).

<div style="overflow: auto;">
  <iframe src="/files/barnes-hut/barnes-hut.html" style="width: 800px; height: 800px; border: none; margin: 0 auto;"></iframe>
</div>

## Other use cases

Since the space is partitioned in an esay to use data structure, we can use it for other things, like collision detection. or serching for objects by position. And basically any other problem where you want to avoid looping over all the objects and simplify computations by grouping objects that are far away.

## References

- Original paper: [A Hierarchical O(N log N) Force Calculation Algorithm](https://ui.adsabs.harvard.edu/abs/1986Natur.324..446B/abstract)
- [Wikipedia](https://en.wikipedia.org/wiki/Barnes%E2%80%93Hut_simulation)
- Barnes hut in action: [The Barnes-Hut Approximation](https://jheer.github.io/barnes-hut) by Jeff Heer
- 3D online simulation: [Barnes-Hut Simulation](https://jurasic.dev/2023/barnes-hut-simulation/) by Jurasic Dev
