---
title: AI activation functions
description: Comparison of activation functions used in neural networks. Activation functions are used to introduce non-linearity into neural networks.
categories:
  - ai
  - computer-science
  - machine-learning
listed: true
---

<script>
  import Sigmoid from './ai-activation-functions/Sigmoid.svelte';
  import ReLu from './ai-activation-functions/ReLu.svelte';
  import LeakyReLu from './ai-activation-functions/LeakyReLu.svelte';
  import DerivativeReLu from './ai-activation-functions/DerivativeReLu.svelte';
  import TanH from './ai-activation-functions/TanH.svelte';
</script>

## Sigmoid

<Sigmoid />

### Mathematical properties

- Continuous and differentiable on $\mathbb{R}$
- Output range of $[0, 1]$

### Applications

Commonly used in the output layer of a binary classifier.
It is also sensitive to changes in the input, which makes it useful for backpropagation.

### Pros

- Smooth gradient, which helps in stable training
- Outputs are in a bounded range, making it suitable for certain types of problems

### Cons

- Vanishing gradient problem: Gradients become very small for extreme inputs, which can slow down training
- Outputs are not centered around zero, which can lead to slower convergence in some cases

## ReLu (_Rectified Linear Unit_)

<ReLu />

### Mathematical properties

- Continuous on $\mathbb{R}$ and differentiable on $\mathbb{R\setminus\{0\}}$
- Output range of $[0, 1]$

### Applications

Widely used in hidden layers of deep neural networks. It addresses the vanishing gradient problem to some extent by preventing gradients from becoming too small.

### Pros

- Dead ReLU problem: Neurons can get stuck and never activate if their weights are not updated properly.
- Not zero-centered, which can lead to gradient issues in some cases.

### Cons

- Dead ReLU problem: Neurons can get stuck and never activate if their weights are not updated properly.
- Not zero-centered, which can lead to gradient issues in some cases.

## Leaky ReLu

<LeakyReLu />

### Mathematical properties

- Continuous on $\mathbb{R}$ and differentiable on $\mathbb{R}\setminus\{0\}$
- Similar to ReLU for $x>0$, but with a small gradient for $x<0$
- Output range of $[-\lambda,1]$

### Applications

Leaky ReLU addresses the dead ReLU problem by allowing a small gradient for negative inputs.

### Pros

- Addresses the dead ReLU problem
- Efficient and easy to implement

### Cons

- The choice of $\lambda$ is important and can affect performance
- Not zero-centered

## Derivative ReLu _or_ Hard Threshold

<DerivativeReLu />

### Mathematical properties

- Continuous on $\mathbb{R}$ and differentiable on $\mathbb{R}\setminus\{0\}$
- Output range of ${0,1}$

### Applications

Derivative ReLU, also known as Hard Threshold, is the simplest activation function. It has only two possible outputs: 0 or 1. So it is used in some contexts where a binary output is desired.

### Pros

- Addresses the dead ReLU problem
- Efficient and easy to implement

### Cons

- The choice of $\lambda$ is important and can affect performance
- Not zero-centered

## TanH

<TanH />

### Mathematical properties

- Continuous and differentiable on $\mathbb{R}$
- Output range of $[-1, 1]$

### Applications

Used in some contexts where zero-centered outputs are desired.

### Pros

- Zero-centered, which can help with faster convergence in some cases
- Smooth gradient for most inputs

### Cons

- Can still suffer from vanishing gradient for very large inputs
- Outputs are bound between -1 and 1, which may not be suitable for all problems
