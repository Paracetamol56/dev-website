// See https://kit.svelte.dev/docs/types#app
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		// interface Locals {}
		interface PageData {
			shell?: Partial<import('$lib/shell').Shell>;
		}
		// interface Platform {}
	}
}

export {};
