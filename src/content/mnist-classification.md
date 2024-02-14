---
title: MNIST Classification
description: A simple MNIST classification example using TensorFlow.js
tags:
  - ai
  - computer-science
  - machine-learning
listed: true
---

<script>
  import { writable } from 'svelte/store';
	import EvaluateModel from './mnist-classification/EvaluateModel.svelte';
	import InitModel from './mnist-classification/InitModel.svelte';
	import LoadDataset from './mnist-classification/LoadDataset.svelte';
	import SaveLoadModel from './mnist-classification/SaveLoadModel.svelte';
	import TestModel from './mnist-classification/TestModel.svelte';
	import TrainModel from './mnist-classification/TrainModel.svelte';
	import { MnistData } from './mnist-classification/mnistData';
	import * as tf from '@tensorflow/tfjs';

	const data = writable(new MnistData());
	const model = writable(tf.sequential());
</script>

## What is MNIST?

MNIST is a dataset of handwritten digits. It is a popular dataset for machine learning because it is relatively small and easy to work with. The dataset contains 60,000 training images and 10,000 testing images. Each image is a 28x28 grayscale image of a handwritten digit. The images are labeled with the digit they represent.

![MNIST Classification](/img/mnist-examples.png)

## Let's build a model !

To build this page, I followed the steps explained in this [Medium article](https://medium.com/ailab-telu/learn-and-play-with-tensorflow-js-part-3-dd31fcab4c4b) by [Ari Pratama](https://medium.com/@undeed).

### Load the MNIST Dataset

First, we need to download the MNIST dataset from the internet. The dataset is provided by Google and is made of 2 parts: [the images](https://storage.googleapis.com/learnjs-data/model-builder/mnist_images.png) and [the labels](https://storage.googleapis.com/learnjs-data/model-builder/mnist_labels_uint8). We also need to preprocess the raw data to store 28x28 images and their corresponding labels in the memory.
After that, we can split the data into training and testing sets.

<LoadDataset {data} />

### Initialize the Model

I defined 2 different models: a simple ANN (Artificial Neural Network) and a CNN (Convolutional Neural Network).
They work similarly, but the CNN applies convolutional layers to the input data in order to extract relevant features from the images.

<InitModel {model} />

### Train the Model

This is where the magic happens. We will train the model using the training set.
The 2 graphs below show the loss and accuracy of the model during the training process.

<TrainModel {data} {model} />

### Evaluate the Model

In order to measure the performance of the model, we can evaluate it using the testing set (which the model has never seen before).
Here you can see the class accuracy showing how well the model performs for each digit. And the confusion matrix showing the number of correct and incorrect predictions. These metrics are very useful to quickly and intuitively understand the performance of the model and compare it against other models.

<EvaluateModel {data} {model} />

### Save or Load the Model

If you want to save the model locally to use it later or to share it with others, you can use the following section.

<SaveLoadModel {model} />

### Test the Model

Finally, we can test the model using a custom image. You can draw a digit in the canvas and see how well the model predicts it.

<TestModel {model} />
