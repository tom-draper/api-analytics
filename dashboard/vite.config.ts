import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig, loadEnv } from 'vite';
import { readFileSync } from 'node:fs';

const { version } = JSON.parse(readFileSync('package.json', 'utf-8'));

export default defineConfig(({ mode }) => {
	// Load all env vars (from .env files and the shell) without requiring a VITE_ prefix.
	const env = loadEnv(mode, process.cwd(), '');
	// Backend URL for a self-hosted dashboard, baked in at build time. Falls back to
	// the legacy VITE_SERVER_URL name, then to '' (meaning: use the default in consts.ts).
	const serverURL = env.SERVER_URL || env.VITE_SERVER_URL || '';

	return {
		plugins: [tailwindcss(), sveltekit()],
		define: {
			__APP_VERSION__: JSON.stringify(version),
			__SERVER_URL__: JSON.stringify(serverURL)
		},
		resolve: {
			alias: {
				$lib: '/src/lib',
				$components: '/src/lib/components',
				$img: '/src/lib/img'
			}
		}
	};
});
