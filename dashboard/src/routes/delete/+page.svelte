<script lang="ts">
	import { getServerURL } from '$lib/url';
	import Lightning from '$components/Lightning.svelte';

	type State = 'idle' | 'loading' | 'deleted' | 'error';

	let state: State = $state('idle');
	let apiKey = $state('');

	async function submit() {
		if (!apiKey) return;

		state = 'loading';

		try {
			const url = getServerURL();
			const response = await fetch(`${url}/api/delete/${apiKey}`);
			if (response.status === 200) {
				apiKey = '';
				state = 'deleted';
			} else {
				state = 'error';
			}
		} catch (e) {
			console.log(e);
			state = 'error';
		}
	}

	function enter(e: KeyboardEvent) {
		if (e.key === 'Enter') submit();
	}

	function reset() {
		state = 'idle';
		apiKey = '';
	}
</script>

<svelte:head>
	<link rel="icon" href="/images/logos/lightning-red.svg" />
</svelte:head>

<div class="form-page">
	<div class="content">
		<div class="logo-icon">
			<Lightning />
		</div>
		<h2 class="title">Delete Account</h2>
		<p class="subtitle">Permanently delete all data associated with your<br>API key.</p>

		<label class="input-label" for="api-key">
			Enter API Key
			<svg class="arrow" viewBox="240 170 320 400" fill="none" xmlns="http://www.w3.org/2000/svg">
				<g stroke-width="31" stroke="currentColor" stroke-linecap="square" transform="matrix(1,0,0,1,-4,0)">
					<path d="M250 256.4Q413 180.4 550 556.4" marker-end="url(#arrowhead-del)"/>
				</g>
				<defs>
					<marker markerWidth="6" markerHeight="6" refX="3" refY="3" viewBox="0 0 6 6" orient="auto" id="arrowhead-del">
						<polygon points="0,6 0,0 6,3" fill="currentColor"/>
					</marker>
				</defs>
			</svg>
		</label>
		<input
			id="api-key"
			type="text"
			bind:value={apiKey}
			placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
			onkeydown={enter}
			disabled={state === 'loading' || state === 'deleted'}
		/>

		{#if state === 'deleted'}
			<div class="status-msg success">
				Your account and all associated data has been deleted.
			</div>
		{:else if state === 'error'}
			<div class="status-msg error">
				Something went wrong. Please check your API key and try again.
			</div>
			<button class="form-btn delete-btn" onclick={reset}>Try again</button>
		{:else}
			<button class="form-btn delete-btn" onclick={submit} disabled={state === 'loading'}>
				{#if state === 'loading'}
					<div class="loader"></div>
				{:else}
					Delete
				{/if}
			</button>
		{/if}
	</div>
</div>

<style scoped>
	.content {
		box-shadow: 0 0 180px 2px var(--red);
		max-width: calc(40ch + 6em);
	}
	.logo-icon {
		color: var(--red);
	}
	.title {
		color: var(--red);
	}
	.delete-btn {
		background: var(--red);
	}
	.delete-btn:hover:not(:disabled) {
		background: #c94f4f;
	}
	.status-msg {
		text-align: center;
	}
	.status-msg.success {
		background: #0d2b1a;
		color: var(--highlight);
		border: 1px solid #1a4a2e;
	}
	.loader {
		border-top-color: var(--red);
	}
</style>
