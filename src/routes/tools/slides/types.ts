import type { DeckSettings } from '$lib/components/deck/settings';

export type OutlineEntry = {
	title: string;
	from: number;
	to: number;
	vertical: number;
};

export type DeckSummary = {
	id: string;
	title: string;
	outline: OutlineEntry[];
	createdAt: string;
	updatedAt: string;
};

export type Deck = DeckSummary & {
	user: string;
	content: string;
	settings: DeckSettings;
	revision: number;
};

export type DeckPage = {
	items: DeckSummary[];
	total: number;
	page: number;
	limit: number;
};

export type Indices = { h: number; v: number; f?: number };

export const perPage = 12;
