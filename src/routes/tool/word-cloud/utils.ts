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
  user: string;
  words: WordCloudWord[];
  createdAt: Date;
  closedAt: Date | null;
}

interface WordCloudSessionUser extends WordCloudSession {
  uuid: string;
}

interface ManagePageData extends WordCloudSession {
  submitions: number,
}

export type { WordCloudWord, WordCloudSessionAdmin, WordCloudSessionUser, ManagePageData };

