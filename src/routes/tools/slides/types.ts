import type { DeckSettings } from '$lib/components/deck/settings';

export type OutlineEntry = {
	title: string;
	from: number;
	to: number;
	vertical: number;
};

export type Role = 'owner' | 'editor' | 'viewer';
export type LinkRole = Exclude<Role, 'owner'>;

export type DeckSummary = {
	id: string;
	role: Role;
	ownerName?: string;
	title: string;
	outline: OutlineEntry[];
	createdAt: string;
	updatedAt: string;
};

// Secret tokens of the share links, only given to the owner
export type Links = Partial<Record<LinkRole, string>>;

export type Member = { user: string; name: string; role: LinkRole; joinedAt: string };

export type Deck = DeckSummary & {
	user: string;
	content: string;
	settings: DeckSettings;
	revision: number;
	links?: Links;
	members?: Member[];
};

export type DeckPage = {
	items: DeckSummary[];
	total: number;
	page: number;
	limit: number;
};

export type Indices = { h: number; v: number; f?: number };

export const perPage = 12;
