const initDrawCanvas = (canvas: HTMLCanvasElement) => {
	const ctx = canvas.getContext('2d');
	if (!ctx) {
		return;
	}
	ctx.fillStyle = '#000000';
	ctx.fillRect(0, 0, canvas.width, canvas.height);
	ctx.lineWidth = 10;
	ctx.lineCap = 'round';
	ctx.strokeStyle = '#ffffff';

	let isDrawing = false;
	let lastX = 0;
	let lastY = 0;

	const draw = (e: MouseEvent) => {
		if (!isDrawing) return;
		ctx.beginPath();
		ctx.moveTo(lastX, lastY);
		ctx.lineTo(e.offsetX, e.offsetY);
		ctx.stroke();
		[lastX, lastY] = [e.offsetX, e.offsetY];
	};

	canvas.addEventListener('mousedown', (e) => {
		isDrawing = true;
		[lastX, lastY] = [e.offsetX, e.offsetY];
	});

	canvas.addEventListener('mousemove', draw);
	canvas.addEventListener('mouseup', () => (isDrawing = false));
	canvas.addEventListener('mouseout', () => (isDrawing = false));
};

const clearCanvas = (canvas: HTMLCanvasElement) => {
	const ctx = canvas.getContext('2d');
	if (!ctx) {
		return;
	}
	ctx.fillStyle = '#000000';
	ctx.fillRect(0, 0, canvas.width, canvas.height);
};

export { initDrawCanvas, clearCanvas };
