type Categories = 'sveltekit' | 'svelte';

type Page = {
	title: string;
	slug: string;
	description: string;
	categories: Categories[];
	published: boolean;
};

export type { Categories, Page };
