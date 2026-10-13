type Point = {
	x: number;
	y: number;
};

type Node = {
	l1: Point;
	l2: Point;
	l3: Point;
	l4: Point;
	opacity: number;
};

function generatePoints(num: number = 100): Point[] {
	return Array.from(
		{ length: num },
		(): Point => ({
			x: Math.random() * 1000,
			y: Math.random() * 1000
		})
	);
}

export { generatePoints };
export type { Point, Node };
