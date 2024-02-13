import { browser } from '$app/environment';
import * as tf from '@tensorflow/tfjs';

async function loadImage(src: string): Promise<tf.Tensor4D> {
	if (!browser) {
		throw new Error('This code must be run in a browser environment');
	}
	let imageTensor: tf.Tensor4D;
	const img: HTMLImageElement = new Image();
	const canvas: HTMLCanvasElement = document.createElement('canvas');
	const ctx: CanvasRenderingContext2D = canvas.getContext('2d')!;

	const imageRequest = new Promise((resolve, reject) => {
		img.src = src;
		img.crossOrigin = 'anonymous';
		img.onload = () => {
			canvas.width = img.naturalWidth;
			canvas.height = img.naturalHeight;
			ctx.drawImage(img, 0, 0);

			const imageData = ctx.getImageData(0, 0, img.naturalWidth, img.naturalHeight);
			imageTensor = tf.browser
				.fromPixels(imageData)
				.expandDims(0)
				.toFloat()
				.div(255)
				.clipByValue(0, 1) as tf.Tensor4D;
			resolve(img);
		};
		img.onerror = reject;
	});

	await imageRequest;
	// @ts-ignore
	return imageTensor;
}

async function applyConvolution(
	matrix: number[][],
	imageTensor: tf.Tensor4D
): Promise<tf.Tensor4D> {
	if (!browser) {
		throw new Error('This code must be run in a browser environment');
	}

	// Create a kernel tensor from the matrix. Shape must be [3, 3, 3, 1]
	const kernel = tf.tensor4d(
		[
			[Array(3).fill([matrix[0][0]]), Array(3).fill([matrix[0][1]]), Array(3).fill([matrix[0][2]])],
			[Array(3).fill([matrix[1][0]]), Array(3).fill([matrix[1][1]]), Array(3).fill([matrix[1][2]])],
			[Array(3).fill([matrix[2][0]]), Array(3).fill([matrix[2][1]]), Array(3).fill([matrix[2][2]])]
		],
		[3, 3, 3, 1]
	);

	const convolvedImageTensor = tf.tidy(() => {
		return tf.depthwiseConv2d(imageTensor, kernel, 1, 'valid').clipByValue(0, 1) as tf.Tensor4D;
	});

	return convolvedImageTensor;
}

async function applyMaxPooling(
	poolSize: number,
	stride: number,
	imageTensor: tf.Tensor4D
): Promise<tf.Tensor4D> {
	if (!browser) {
		throw new Error('This code must be run in a browser environment');
	}

	const pooledImageTensor = tf.tidy(() => {
		return tf.maxPool(imageTensor, poolSize, stride, 'valid') as tf.Tensor4D;
	});

	return pooledImageTensor;
}

export { loadImage, applyConvolution, applyMaxPooling };
