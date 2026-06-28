
export default interface SlideDeck {
  id: string;
  title: string;
  content: string;
  user_id: string;
  revision: number;
  status: 'draft' | 'private' | 'published';
  created_at: Date;
  updated_at: Date
}
