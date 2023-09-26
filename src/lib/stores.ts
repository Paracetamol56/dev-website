import type { ObjectId } from 'mongodb';
import { writable, type Writable } from 'svelte/store';

type User = {
  _id: ObjectId;
  name: string|null;
  email: string;
};
const user: Writable<User|null> = writable(null);

export { user };