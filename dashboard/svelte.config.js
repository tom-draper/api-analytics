import adapterAuto from '@sveltejs/adapter-auto';
import adapterNode from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

// adapter-auto powers the hosted deploy; set ADAPTER=node to produce a standalone
// Node server (`node build`) for self-hosting via the Dockerfile.
const adapter = process.env.ADAPTER === 'node' ? adapterNode : adapterAuto;

/** @type {import('@sveltejs/kit').Config} */
const config = {
	// Consult https://svelte.dev/docs/kit/integrations
	// for more information about preprocessors
	preprocess: vitePreprocess(),

	kit: {
		adapter: adapter(),
		alias: {
			$components: './src/lib/components',
			'$components/*': './src/lib/components/*'
		}
	}
};

export default config;
