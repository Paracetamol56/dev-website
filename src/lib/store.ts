import { browser } from '$app/environment';
import type { variants } from '@catppuccin/palette';
import { writable, type Writable } from 'svelte/store';

// User store
type UserStore = {
	accessToken: string | null;
	refreshToken: string | null;
	id: string | null;
	email: string | null;
	flavour: keyof typeof variants;
};
const user: Writable<UserStore> = initUserStore();

function initUserStore(): Writable<UserStore> {
	let value: UserStore = {
		id: null,
		accessToken: null,
		refreshToken: null,
		email: null,
		flavour: 'mocha'
	};
	if (browser) {
		const cookie = document.cookie.split('; ').find((row) => row.startsWith('user='));
		if (cookie) {
			const { id, accessToken, refreshToken, email, flavour } = JSON.parse(cookie.split('=')[1]);
			value = {
				id: id || null,
				accessToken: accessToken || null,
				refreshToken: refreshToken || null,
				email: email || null,
				flavour: flavour || 'mocha'
			};
			applyUserTheme(flavour || 'mocha');
		}
	}

	writeUserToCookie(value);
	return writable(value);
}

// Overide the user store's set method to write to the cookie as well
const { set: userSet } = user;
user.set = (value) => {
	applyUserTheme(value.flavour);
	writeUserToCookie(value);
	userSet(value);
};

// Function to write the user store to the cookie
function writeUserToCookie(user: UserStore) {
	if (browser) {
		// Delete existing cookie
		document.cookie = 'user=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/; SameSite=Strict;';
		// Set new cookie
		document.cookie = `user=${JSON.stringify(user)}; path=/; SameSite=Strict;`;
	}
}

function applyUserTheme(theme: keyof typeof variants) {
	if (!browser) return;
	const body = document.querySelector('body');
	body?.classList.forEach((className) => {
		if (className.startsWith('ctp-')) {
			body?.classList.remove(className);
		}
	});
	body?.classList.add(`ctp-${theme}`);
}

export { user };
