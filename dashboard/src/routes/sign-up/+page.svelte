<script lang="ts">
	import { getServerURL } from '$lib/url';
	import Lightning from '$components/Lightning.svelte';

	type State = 'generate' | 'loading' | 'copy' | 'copied' | 'error';

	let apiKey = $state('');
	let state = $state<State>('generate');

	async function submit() {
		if (apiKey) return;

		state = 'loading';

		try {
			const url = getServerURL();
			const response = await fetch(`${url}/api/generate-api-key`);
			if (response.status === 200) {
				apiKey = await response.json();
				state = 'copy';
			} else {
				state = 'error';
			}
		} catch (e) {
			console.log(e);
			state = 'generate';
		}
	}

	function copyToClipboard() {
		navigator.clipboard.writeText(apiKey);
		state = 'copied';
	}
</script>

<div class="form-page">
	<div class="content">
		<div class="logo-icon">
			<Lightning />
		</div>
		<h2 class="title">Generate API Key</h2>
		<p class="subtitle">Get a free API key to start tracking your requests</p>

		<label class="input-label" class:label-ready={state === 'copy' || state === 'copied'} for="api-key">
			Your API Key
			<svg class="arrow" viewBox="240 170 320 400" fill="none" xmlns="http://www.w3.org/2000/svg">
				<g stroke-width="31" stroke="currentColor" stroke-linecap="square" transform="matrix(1,0,0,1,-4,0)">
					<path d="M250 256.4Q413 180.4 550 556.4" marker-end="url(#arrowhead-gen)"/>
				</g>
				<defs>
					<marker markerWidth="6" markerHeight="6" refX="3" refY="3" viewBox="0 0 6 6" orient="auto" id="arrowhead-gen">
						<polygon points="0,6 0,0 6,3" fill="currentColor"/>
					</marker>
				</defs>
			</svg>
		</label>
		<input
			id="api-key"
			type="text"
			readonly
			bind:value={apiKey}
			placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
			class:input-ready={state === 'copy' || state === 'copied'}
		/>

		{#if state === 'error'}
			<p class="error-msg">Something went wrong. Please try again.</p>
		{/if}

		{#if state === 'loading'}
			<button class="form-btn" disabled aria-label="Loading">
				<div class="loader"></div>
			</button>
		{:else if state === 'copy' || state === 'copied'}
			<button class="form-btn copy-btn" onclick={copyToClipboard}>
				{state === 'copied' ? 'Copied ✓' : 'Copy to clipboard'}
			</button>
		{:else}
			<button class="form-btn" onclick={submit}>Generate</button>
		{/if}

		<p class="keep-safe" class:keep-safe-visible={state === 'copy' || state === 'copied'}>
			Keep your API key safe – it grants access to your analytics data.
		</p>
	</div>
</div>

<style scoped>
	.logo-icon {
		color: var(--highlight);
	}
	.title {
		color: var(--highlight);
	}
	.input-label {
		opacity: 0;
		transition: opacity 0.4s;
	}
	.label-ready {
		opacity: 1;
	}
	.form-btn {
		background: #3fcf8e;
	}
	input {
		color: var(--muted-text);
		transition: color 0.2s;
	}
	.input-ready {
		color: white !important;
	}
	.copy-btn:hover {
		background: #2ea872 !important;
	}
	.error-msg {
		font-size: 0.8em;
		color: var(--red);
		text-align: left;
		margin-top: -1.8em;
		margin-bottom: 1.2em;
		padding: 0;
	}
	.keep-safe {
		font-size: 0.78em;
		color: var(--dim-text);
		padding: 0;
		max-height: 0;
		overflow: hidden;
		opacity: 0;
		transition: max-height 0.4s ease, opacity 0.4s, padding-top 0.4s;
	}
	.keep-safe-visible {
		max-height: 3em;
		opacity: 1;
		padding-top: 1.5em;
	}
</style>
