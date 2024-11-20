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
  open: boolean;
}

interface WordCloudSessionAdmin extends WordCloudSession {
  user: string;
  words: WordCloudWord[];
  createdAt: Date;
  closedAt: Date | null;
}

interface WordCloudSessionUser extends WordCloudSession {
  uuid: string;
}

export type { WordCloudWord, WordCloudSessionAdmin, WordCloudSessionUser };
