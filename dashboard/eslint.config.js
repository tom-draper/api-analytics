import prettier from 'eslint-config-prettier';
import js from '@eslint/js';
import { includeIgnoreFile } from '@eslint/compat';
import svelte from 'eslint-plugin-svelte';
import globals from 'globals';
import { fileURLToPath } from 'node:url';
import ts from 'typescript-eslint';
const gitignorePath = fileURLToPath(new URL('./.gitignore', import.meta.url));

export default ts.config(
	includeIgnoreFile(gitignorePath),
	// One-off build-time logo generator, not application source.
	{ ignores: ['scripts/'] },
	js.configs.recommended,
	...ts.configs.recommended,
	...svelte.configs['flat/recommended'],
	prettier,
	...svelte.configs['flat/prettier'],
	{
		languageOptions: {
			globals: {
				...globals.browser,
				...globals.node
			}
		}
	},
	{
		files: ['**/*.svelte'],

		languageOptions: {
			parserOptions: {
				parser: ts.parser
			}
		}
	},
	{
		// TypeScript already resolves identifiers (incl. ambient globals from
		// app.d.ts), so the core no-undef rule only produces false positives here.
		files: ['**/*.ts', '**/*.svelte'],
		rules: {
			'no-undef': 'off'
		}
	},
	{
		rules: {
			// Plotly is integrated as an untyped CDN global, so `any` is intentional
			// at those boundaries (chart event payloads, the global itself).
			'@typescript-eslint/no-explicit-any': 'off',
			// Our Set/Map/Date instances are plain caches/computation, not reactive
			// state, so the Svelte reactive variants aren't needed.
			'svelte/prefer-svelte-reactivity': 'off',
			// Links here are mostly external/static; the SvelteKit resolve() guard
			// targets internal programmatic navigation, which isn't the pattern.
			'svelte/no-navigation-without-resolve': 'off'
		}
	},
	{
		// $bindable(false) prop defaults read as "useless" to this rule because it
		// doesn't model Svelte's bindable fallback semantics.
		files: ['**/*.svelte'],
		rules: {
			'no-useless-assignment': 'off'
		}
	}
);
