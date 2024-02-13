---
title: Convolution
description: Demonstration of common convolution operations used in computer vision.
tags:
  - ai
  - computer-science
  - computer-vision
listed: true
---

<script>
	import { ExternalLink, ArrowRight } from 'lucide-svelte';
  import Convolution from './convolution/Convolution.svelte';
  import MaxPooling from './convolution/MaxPooling.svelte';
</script>

## Kernel Convolution

Kernel convolution is a fundamental operation in image processing used for tasks such as blurring, sharpening, edge detection, and more. It involves taking a matrix, or "kernel", and sliding it over the image, pixel by pixel. At each position, the pixel values of the image under the kernel are multiplied by the corresponding values in the kernel, and the results are summed up to give a single output pixel in the resulting image. This process is repeated for every pixel in the image. The choice of kernel determines the effect on the image. The kernel convolution operation can be applied to color images by performing the operation separately on each color channel.

![Animation of kernel convolution](/img/2d-convolution-animation.gif)

<p class="text-center">Animation of kernel convolution. Source: <a class="inline-flex items-baseline" href="https://en.wikipedia.org/wiki/Kernel_(image_processing)" target="_blank">Wikipedia&nbsp;<ExternalLink size="16" /></a></p>

**Parameters**:

- **Kernel**: The matrix used to perform the convolution operation.
- **Stride**: The number of pixels the kernel moves at each step.
- **Padding** _(optional)_: The number of pixels added to the edges of the image to ensure that the kernel can be applied to all pixels.

### Example

Let's consider a simple example with a 3x3 edge detection matrix:

$$\begin{bmatrix} -1 & -1 & -1 \\ -1 & 8 & -1 \\ -1 & -1 & -1 \end{bmatrix}$$

To perform convolution on a single pixel, we place the center of the kernel on the pixel of interest. Then, we multiply the values of the kernel with the corresponding pixel values in the image. Finally, we sum up the results to obtain the output value for that pixel.

For instance, if we have the following pixel values:

$$\begin{bmatrix} 10 & 20 & 30 \\ 40 & 50 & 60 \\ 70 & 80 & 90 \end{bmatrix}$$

and we apply the edge detection kernel to the center pixel (50), the convolution operation would be:

$$(-1 \times 10) \\ + (-1 \times 20) \\ + (-1 \times 30) \\ + (-1 \times 40) \\ + (8 \times 50) \\ + (-1 \times 60) \\ + (-1 \times 70) \\ + (-1 \times 80) \\ + (-1 \times 90) = 50$$

### Larger kernels

Even if 3x3 kernels are the most common, larger kernels can be used to achieve more complex effects. For example, a 5x5 kernel can be used to perform Gaussian blurring:

$$\frac{1}{256} \times \begin{bmatrix} 1 & 4 & 6 & 4 & 1 \\ 4 & 16 & 24 & 16 & 4 \\ 6 & 24 & 36 & 24 & 6 \\ 4 & 16 & 24 & 16 & 4 \\ 1 & 4 & 6 & 4 & 1 \end{bmatrix}$$

<Convolution imgSrc="/img/red-panda.jpg" />

## Max Pooling

Processing images through a neural network can be computationally intensive due to the large number of pixels that act as inputs. An alternative approach is Pooling, which reduces the size of the input by sampling from specific regions within the input. Adjacent pixels typically belong to the same region of the image and are likely to have similar values. A single pixel can be used to represent an entire region. One method to achieve this is through Max-Pooling, where the pixel with the maximum value in a given region is chosen to represent that region.

**Parameters**:

- **Window size**: The size of the region to sample from.
- **Stride**: The number of pixels the window moves at each step.

### Example

In this example, we have a 4x4 input image. We apply a 2x2 max-pooling operation with a stride of 2. The stride is the number of pixels the pooling window moves at each step. In this case, the stride is 2, so the pooling window moves 2 pixels to the right and 2 pixels down at each step. The result is a 2x2 output image.

<div class="flex items-center gap-4">
  <table class="w-32 h-32 table-fixed text-ctp-base font-semibold">
    <tbody>
      <tr>
        <td class="border border-ctp-base text-center bg-ctp-red">12</td>
        <td class="border border-ctp-base text-center bg-ctp-red">20</td>
        <td class="border border-ctp-base text-center bg-ctp-yellow">30</td>
        <td class="border border-ctp-base text-center bg-ctp-yellow">0</td>
      </tr>
      <tr>
        <td class="border border-ctp-base text-center bg-ctp-red">8</td>
        <td class="border border-ctp-base text-center bg-ctp-red">12</td>
        <td class="border border-ctp-base text-center bg-ctp-yellow">2</td>
        <td class="border border-ctp-base text-center bg-ctp-yellow">0</td>
      </tr>
      <tr>
        <td class="border border-ctp-base text-center bg-ctp-blue">34</td>
        <td class="border border-ctp-base text-center bg-ctp-blue">70</td>
        <td class="border border-ctp-base text-center bg-ctp-green">37</td>
        <td class="border border-ctp-base text-center bg-ctp-green">4</td>
      </tr>
      <tr>
        <td class="border border-ctp-base text-center bg-ctp-blue">112</td>
        <td class="border border-ctp-base text-center bg-ctp-blue">100</td>
        <td class="border border-ctp-base text-center bg-ctp-green">25</td>
        <td class="border border-ctp-base text-center bg-ctp-green">12</td>
      </tr>
    </tbody>
  </table>
  <ArrowRight size="24" />
  <table class="w-16 h-16 table-fixed text-ctp-base font-semibold">
    <tbody>
      <tr>
        <td class="border border-ctp-base text-center bg-ctp-red">20</td>
        <td class="border border-ctp-base text-center bg-ctp-yellow">30</td>
      </tr>
      <tr>
        <td class="border border-ctp-base text-center bg-ctp-blue">112</td>
        <td class="border border-ctp-base text-center bg-ctp-green">37</td>
      </tr>
    </tbody>
  </table>
</div>

<MaxPooling imgSrc="/img/red-panda.jpg" />
