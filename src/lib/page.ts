interface Page {
	title: string;
	slug: string;
	description: string;
	tags: string[];
	listed: boolean;
};

interface Tool extends Page {
	needsAuth: boolean;
};

export type { Page, Tool };
