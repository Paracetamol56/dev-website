import { browser } from '$app/environment';
import { variants } from '@catppuccin/palette';
import type { ObjectId } from 'mongodb';
import { writable, type Writable } from 'svelte/store';

// User store
type User = {
	token: string;
	id: ObjectId;
	name: string | null;
	email: string;
	flavour: string;
};
const user: Writable<User | null> = writable(null);

// Overide the user store's set method to write to cookies
const { set: userSet } = user;
user.set = (userObject: User | null) => {
	if (userObject === null) {
		document.cookie = 'user=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/; sameSite=strict';
	} else {
		document.cookie = `user=${JSON.stringify(userObject)}; expires=${new Date(
			Date.now() + 2592000000
		).toUTCString()} path=/; sameSite=strict`;
	}
	userSet(userObject);
};

// Theme store
const theme: Writable<keyof typeof variants> = writable();
theme.set('mocha');

// Overide the theme store's set method to write to cookies
const { set: themeSet } = theme;
theme.set = (themeName: string) => {
	// Validate
	if (!Object.keys(variants).includes(themeName)) {
		throw new Error(`Invalid theme: ${themeName}`);
	}
	// Expiration: 30 days
	document.cookie = `theme=${themeName}; expires=${new Date(
		Date.now() + 2592000000
	).toUTCString()} path=/; sameSite=strict`;
	if (browser) {
		const body = document.querySelector('body');
		body?.classList.forEach((className) => {
			if (className.startsWith('ctp-')) {
				body?.classList.remove(className);
			}
		});
		body?.classList.add(`ctp-${themeName}`);
	}
	themeSet(themeName as keyof typeof variants);
};

const readCookieAndSetStore = <T>(cookieName: string, store: Writable<T>) => {
	const cookie = document.cookie.split('; ').find((item) => item.startsWith(cookieName));
	if (cookie) {
		try {
			store.set(JSON.parse(cookie.split('=')[1]));
		} catch (e) {
			console.error(e);
		}
	}
};

// Read persistant data from cookies
if (browser) {
	// User cookie
	readCookieAndSetStore<User | null>('user', user);
	// Theme cookie
	const cookie = document.cookie.split('; ').find((item) => item.startsWith('theme'));
	if (cookie) {
		try {
			theme.set(cookie.split('=')[1] as keyof typeof variants);
		} catch (e) {
			console.error(e);
		}
	}
}

export { user, theme };
