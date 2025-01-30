<script lang="ts">
	import { getServerURL } from '../lib/url';

	type State = 'generate' | 'loading' | 'copy' | 'copied' | 'error';

	let state: State = 'generate';
	let apiKey = '';
	async function submit() {
		if (apiKey) {
			return; // Already generated
		}

		setState('loading');

		try {
			const url = getServerURL();
			const response = await fetch(`${url}/api/generate-api-key`);
			if (response.status === 200) {
				const data = await response.json();
				apiKey = data;
				setState('copy');
			} else {
				setState('error');
			}
		} catch (e) {
			console.log(e);
			setState('generate');
		}
	}

	function setState(value: State) {
		state = value;
	}

	function enter(e) {
		if (e.keyCode === 13) {
			submit();
		}
	}

	function copyToClipboard() {
		navigator.clipboard.writeText(apiKey);
		setState('copied');
	}
</script>

<div class="generate">
	<div class="content">
		<h2>Generate API key</h2>
		<input type="text" readonly bind:value={apiKey} on:keydown={enter} />
		<button
			id="formBtn"
			on:click={submit}
			class:no-display={state !== 'generate'}>Generate</button
		>
		<button id="formBtn" class:no-display={state !== 'loading'}>
			<div class="spinner">
				<div class="loader" />
			</div>
		</button>
		<button
			id="formBtn"
			on:click={copyToClipboard}
			class:no-display={state !== 'copy'}
			><img class="copy-icon" src="img/icons/copy.png" alt="" /></button
		>
		<button
			id="formBtn"
			class="copied-btn"
			on:click={copyToClipboard}
			class:no-display={state !== 'copied'}>Copied</button
		>
		<button id="formBtn" class:no-display={state != 'error'}>Error</button>
	</div>
	<div class="details">
		<div class="keep-secure">Keep your API key safe and secure.</div>
		<div class="highlight logo">API Analytics</div>
		<img class="footer-logo" src="img/logos/lightning-green.png" alt="" />
	</div>
</div>

<style scoped>
	.copy-icon {
		height: 18px;
	}
	.copied-btn {
		cursor: default;
	}
	.spinner {
		height: auto;
	}
	.loader {
		border: 3px solid #343434;
		border-top: 3px solid var(--highlight);
		height: 10px;
		width: 10px;
	}
</style>
