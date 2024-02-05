---
title: MNIST Classification
description: A simple MNIST classification example using TensorFlow.js
categories:
  - ai
  - computer-science
  - machine-learning
listed: true
---

<script>
  import MnistClassification from './mnist-classification/MnistClassification.svelte';
</script>

## What is MNIST?

MNIST is a dataset of handwritten digits. It is a popular dataset for machine learning because it is relatively small and easy to work with. The dataset contains 60,000 training images and 10,000 testing images. Each image is a 28x28 grayscale image of a handwritten digit. The images are labeled with the digit they represent.

![MNIST Classification](/img/mnist-examples.png)

To build this page, I followed the steps explained in this [Medium article](https://medium.com/ailab-telu/learn-and-play-with-tensorflow-js-part-3-dd31fcab4c4b) by [Ari Pratama](https://medium.com/@undeed).

<MnistClassification />
