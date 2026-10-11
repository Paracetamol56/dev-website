import { sveltekit } from '@sveltejs/kit/vite';
import { sveltekitOG } from '@ethercorps/sveltekit-og/plugin';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit(), sveltekitOG()],
	server: {
		proxy:
			process.env.NODE_ENV === 'production'
				? undefined
				: {
						'/api': {
							target: 'http://127.0.0.1:8080',
							changeOrigin: true,
							ws: true
						}
				  }
	}
});
