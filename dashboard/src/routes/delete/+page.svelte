<script lang="ts">
	import { getServerURL } from '$lib/url';
	import Lightning from '$components/Lightning.svelte';

	type State = 'idle' | 'confirm' | 'loading' | 'deleted' | 'error';

	let status: State = $state('idle');
	let apiKey = $state('');

	// First step: ask for explicit confirmation rather than deleting straight
	// away. Deletion is irreversible, so it must never happen on a single action.
	function requestDelete() {
		if (!apiKey || status === 'loading') return;
		status = 'confirm';
	}

	// Second step: the actual deletion, only reachable from an explicit confirm.
	async function confirmDelete() {
		if (!apiKey) return;

		status = 'loading';

		try {
			const url = getServerURL();
			// Sent as a POST body to keep the API key out of URLs and server logs.
			// No Content-Type is set on purpose: that keeps this a CORS simple
			// request, avoiding a preflight the server does not answer. The other
			// POST endpoints do the same.
			const response = await fetch(`${url}/api/delete`, {
				method: 'POST',
				body: JSON.stringify({ api_key: apiKey })
			});
			if (response.status === 200) {
				apiKey = '';
				status = 'deleted';
			} else {
				status = 'error';
			}
		} catch (e) {
			console.log(e);
			status = 'error';
		}
	}

	// Enter only advances to the confirmation step; it never performs the
	// deletion, so a stray keypress cannot delete an account.
	function enter(e: KeyboardEvent) {
		if (e.key === 'Enter' && status === 'idle') requestDelete();
	}

	function cancel() {
		status = 'idle';
	}

	function reset() {
		status = 'idle';
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
		<p class="subtitle">Permanently delete all data associated with your<br />API key.</p>

		<label class="input-label" for="api-key">
			Enter API Key
			<svg
				class="arrow"
				viewBox="240 170 320 400"
				fill="none"
				xmlns="http://www.w3.org/2000/svg"
			>
				<g
					stroke-width="31"
					stroke="currentColor"
					stroke-linecap="square"
					transform="matrix(1,0,0,1,-4,0)"
				>
					<path d="M250 256.4Q413 180.4 550 556.4" marker-end="url(#arrowhead-del)" />
				</g>
				<defs>
					<marker
						markerWidth="6"
						markerHeight="6"
						refX="3"
						refY="3"
						viewBox="0 0 6 6"
						orient="auto"
						id="arrowhead-del"
					>
						<polygon points="0,6 0,0 6,3" fill="currentColor" />
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
			disabled={status === 'confirm' || status === 'loading' || status === 'deleted'}
		/>

		{#if status === 'deleted'}
			<div class="status-msg success">
				Your account and all associated data has been deleted.
			</div>
		{:else if status === 'error'}
			<div class="status-msg error">
				Something went wrong. Please check your API key and try again.
			</div>
			<button class="form-btn delete-btn" onclick={reset}>Try again</button>
		{:else if status === 'confirm'}
			<div class="status-msg warning">
				This permanently deletes your account and all associated data. This action cannot be
				undone.
			</div>
			<div class="confirm-actions">
				<button class="form-btn cancel-btn" onclick={cancel}>Cancel</button>
				<button class="form-btn delete-btn" onclick={confirmDelete}
					>Yes, delete everything</button
				>
			</div>
		{:else}
			<button
				class="form-btn delete-btn"
				onclick={requestDelete}
				disabled={status === 'loading'}
			>
				{#if status === 'loading'}
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
	.status-msg.warning {
		background: #2b0d0d;
		color: var(--red);
		border: 1px solid #4a1a1a;
	}
	.confirm-actions {
		display: flex;
		gap: 0.75em;
	}
	.confirm-actions .form-btn {
		flex: 1;
	}
	.cancel-btn {
		background: #333333;
		color: white;
	}
	.cancel-btn:hover:not(:disabled) {
		background: #444444;
	}
	.loader {
		border-top-color: var(--red);
	}
</style>
