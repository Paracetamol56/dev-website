import type { Component } from 'svelte';

interface Content {
	title: string;
	slug: string;
	description: string;
	tags: string[];
	listed: boolean;
}

interface Page extends Content {
	release: Date;
}

interface Tool extends Content {
	order: number;
	requiresAuth: boolean;
	content?: Component;
}

export type { Page, Tool };
