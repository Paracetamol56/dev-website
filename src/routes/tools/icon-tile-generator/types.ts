export type IconSource = 'lucide' | 'simpleicons';

export type Icon = {
	id: string;
	name: string;
	title: string;
	source: IconSource;
	tags: string[];
	hex?: string;
};

export const SHAPES = ['square', 'round'] as const;
export type Shape = (typeof SHAPES)[number];

export const FORMATS = ['png', 'jpg', 'webp', 'svg', 'bmp', 'ico'] as const;
export type Format = (typeof FORMATS)[number];
