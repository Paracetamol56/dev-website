interface Page {
	title: string;
	slug: string;
	description: string;
	tags: string[];
	release: Date;
	listed: boolean;
}

interface Tool extends Page {
	requiresAuth: boolean;
}

export type { Page, Tool };
