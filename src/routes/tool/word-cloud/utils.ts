interface WordCloudWord {
	text: string;
	uuid: string;
	userAgent: string;
	createdAt: Date;
}

interface WordCloudSession {
	id: string;
	name: string;
	description: string;
	code: string;
}

interface WordCloudSessionAdmin extends WordCloudSession {
	open: boolean;
	user: string;
	words: WordCloudWord[];
	createdAt: Date;
	closedAt: Date | null;
}

interface WordCloudSessionUser extends WordCloudSession {
	open: boolean;
}

interface ManagePageData extends WordCloudSession {
	submissions: number;
	open: boolean;
	createdAt: string;
	closedAt: string | null;
}

export type { WordCloudWord, WordCloudSessionAdmin, WordCloudSessionUser, ManagePageData };
