
interface WordCloudWord {
  text: string,
  ip: string,
  userAgent: string,
  createdAt: Date,
}

interface WordCloudSession {
  id: string,
  user: string,
  name: string,
  description: string,
  code: string,
  open: boolean,
  words: WordCloudWord[],
  createdAt: Date,
  closedAt: Date | null,
}

export type { WordCloudWord, WordCloudSession };
