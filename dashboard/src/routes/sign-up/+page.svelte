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

<div class="generate">
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
		height: 36px;
		margin: 0 auto 1em;
	}
	.title {
		font-size: 2em;
		font-weight: 700;
		color: var(--highlight);
		margin-bottom: 0.2em;
	}
	.subtitle {
		font-size: 0.9em;
		color: var(--dim-text);
		margin-bottom: 2em;
		padding: 0;
	}
	.input-label {
		display: block;
		text-align: left;
		font-size: 0.8em;
		color: var(--dim-text);
		margin-bottom: 0.5em;
		opacity: 0;
		transition: opacity 0.4s;
	}
	.label-ready {
		opacity: 1;
	}
	.arrow {
		display: inline-block;
		width: 12px;
		height: 17px;
		margin-left: 3px;
		vertical-align: middle;
	}
	.form-btn {
		font-size: 0.9em;
		height: 40px;
		border-radius: 4px;
		padding: 0 20px;
		border: none;
		cursor: pointer;
		width: auto;
		background: #3fcf8e;
		font-family: 'Noto Sans', 'Geist' !important;
		font-weight: 400;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}
	input {
		color: #505050;
		transition: color 0.2s;
	}
	input::placeholder {
		color: #707070;
	}
	.input-ready {
		color: white !important;
	}
	.copy-btn {
		width: auto;
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
	.loader {
		border: 3px solid #343434;
		border-top: 3px solid var(--highlight);
		width: 1em;
		height: 1em;
	}
</style>
