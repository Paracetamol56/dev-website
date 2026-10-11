interface WordCloudWord {
	id: string;
	text: string;
	originalText?: string;
	hidden: boolean;
	uuid: string;
	userAgent: string;
	createdAt: string;
	/** Keyed hashes of the browser details and of the IP address */
	device?: string;
	network?: string;
	/** From -5 (very negative) to 5 (very positive), when the word is rated */
	sentiment?: number;
}

interface Settings {
	showCloudToAudience: boolean;
	filterProfanity: boolean;
	allowUpvotes: boolean;
	sentiment: boolean;
	/** Show the mood in presenter mode, when sentiment analysis is on */
	sentimentInPresenter: boolean;
	/** Pop up new answers in presenter mode */
	/** Words each participant may send, 0 for no limit; null (unset) means the default */
	maxWords: number | null;
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
	settings: Settings;
	createdAt: string;
	closedAt: string | null;
}

interface WordCloudSessionUser extends WordCloudSession {
	open: boolean;
	settings: Pick<Settings, 'showCloudToAudience' | 'allowUpvotes' | 'maxWords'>;
	/** Visible words, when the host shows the cloud to the audience */
	words?: string[];
	/** Latest distinct visible words, when upvotes are allowed */
	recent?: string[];
}

interface ManagePageData extends WordCloudSession {
	submissions: number;
	open: boolean;
	createdAt: string;
	closedAt: string | null;
}

const DEFAULT_SETTINGS: Settings = {
	showCloudToAudience: false,
	filterProfanity: true,
	allowUpvotes: true,
	sentiment: true,
	sentimentInPresenter: false,
	maxWords: 20
};

type Distribution = { text: string; occurence: number }[];

/** Occurrences of each text, most frequent first. */
function countWords(texts: string[]): Distribution {
	const counts = new Map<string, number>();
	for (const text of texts) counts.set(text, (counts.get(text) ?? 0) + 1);
	return [...counts]
		.map(([text, occurence]) => ({ text, occurence }))
		.sort((a, b) => b.occurence - a.occurence || a.text.localeCompare(b.text));
}

export { countWords, DEFAULT_SETTINGS };
export type {
	Distribution,
	ManagePageData,
	Settings,
	WordCloudSessionAdmin,
	WordCloudSessionUser,
	WordCloudWord
};
