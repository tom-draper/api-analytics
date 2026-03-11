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
			state = response.status === 200 ? 'deleted' : 'error';
		} catch (e) {
			console.log(e);
			state = 'error';
		}
	}

	function enter(e: KeyboardEvent) {
		if (e.key === 'Enter') submit();
	}
</script>

<div class="generate">
	<div class="content">
		<div class="logo-icon">
			<Lightning />
		</div>
		<h2 class="title">Delete Account</h2>
		<p class="subtitle">Permanently delete all data associated with your API key.</p>

		{#if state === 'deleted'}
			<div class="status-msg success">
				Your account and all associated data has been deleted.
			</div>
		{:else if state === 'error'}
			<div class="status-msg error">
				Something went wrong. Please check your API key and try again.
			</div>
			<button id="formBtn" class="delete-btn" onclick={() => (state = 'idle')}>Try again</button>
		{:else}
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
				disabled={state === 'loading'}
			/>
			{#if state === 'loading'}
				<div id="formBtn" class="grid place-items-center">
					<div class="loader"></div>
				</div>
			{:else}
				<button id="formBtn" class="delete-btn" onclick={submit}>Delete</button>
			{/if}
		{/if}
	</div>
</div>

<style scoped>
	.content {
		box-shadow: 0 0 180px 2px var(--red);
	}
	.logo-icon {
		color: var(--red);
		height: 36px;
		margin: 0 auto 1em;
	}
	.title {
		font-size: 2em;
		font-weight: 700;
		color: var(--red);
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
	}
	.arrow {
		display: inline-block;
		width: 12px;
		height: 17px;
		margin-left: 3px;
		vertical-align: middle;
	}
	#formBtn {
		font-size: 0.9em;
	}
	input::placeholder {
		color: #707070;
	}
	.delete-btn {
		background: var(--red) !important;
	}
	.delete-btn:hover {
		background: #c94f4f !important;
	}
	.status-msg {
		font-size: 0.9em;
		padding: 0.9em 1.2em;
		border-radius: 4px;
		text-align: left;
		line-height: 1.5;
		margin-bottom: 1.2em;
	}
	.status-msg.success {
		background: #0d2b1a;
		color: var(--highlight);
		border: 1px solid #1a4a2e;
	}
	.status-msg.error {
		background: #2b0d0d;
		color: var(--red);
		border: 1px solid #4a1a1a;
	}
	.loader {
		border: 3px solid #343434;
		border-top: 3px solid var(--red);
		width: 1em;
		height: 1em;
	}
</style>
