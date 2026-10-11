import type { variants } from '@catppuccin/palette';

export interface Identity {
	provider: string;
	providerId: string;
	email: string;
	username?: string;
	name?: string;
	avatarUrl?: string;
	profileUrl?: string;
	linkedAt: string;
	lastLogin: string;
}

export interface Passkey {
	id: string;
	name: string;
	createdAt: string;
	lastUsed?: string;
}

export interface UserSettings {
	id: string;
	name: string;
	email: string;
	createdAt: Date;
	lastLogin: Date;
	flavour: keyof typeof variants;
	identities: Identity[] | null;
	passkeys: Passkey[] | null;
}
