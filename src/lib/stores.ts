import type { ObjectId } from 'mongodb';
import { writable, type Writable } from 'svelte/store';

type User = {
  _id: ObjectId;
  name: string|null;
  email: string;
  flavour: string;
};
const user: Writable<User|null> = writable(null);

// Write and read the user from cookies on page load
if (typeof window !== 'undefined') {
  const cookie = document.cookie
    .split('; ')
    .find(row => row.startsWith('user='));
  if (cookie) {
    const userString: string = cookie.split('=')[1];
    const userObject: User = JSON.parse(userString);
    user.set(userObject);
  }
}

// Overide the user store's set method to write to cookies
const { set } = user;
user.set = (userObject: User|null) => {
  if (userObject === null) {
    document.cookie = 'user=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/; sameSite=strict';
  } else {
    document.cookie = `user=${JSON.stringify(userObject)}; path=/; sameSite=strict`;
  }
  set(userObject);
};

export { user };
