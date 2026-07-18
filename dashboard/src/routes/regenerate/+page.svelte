<script lang="ts">
	import { getServerURL, untrustedBackendOrigin } from '$lib/url';
	import Lightning from '$components/Lightning.svelte';

	type State = 'idle' | 'confirm' | 'loading' | 'success' | 'error';

	let status = $state<State>('idle');
	let apiKey = $state('');
	// Set when ?source= points the backend at a non-default origin; the user must
	// confirm before the API key is sent there.
	let destinationWarning = $state<string | null>(null);

	function submit() {
		if (!apiKey || status === 'loading') return;

		const origin = untrustedBackendOrigin();
		if (origin) {
			destinationWarning = origin;
			status = 'confirm';
			return;
		}
		send();
	}

	async function send() {
		status = 'loading';

		try {
			const url = getServerURL();
			// Sent as a POST body (no custom header) to keep the API key out of
			// URLs and to stay a CORS simple request, matching the delete page.
			const response = await fetch(`${url}/api/regenerate-user-id`, {
				method: 'POST',
				body: JSON.stringify({ api_key: apiKey })
			});
			status = response.status === 200 ? 'success' : 'error';
		} catch (e) {
			console.log(e);
			status = 'error';
		}
	}

	function cancel() {
		status = 'idle';
		destinationWarning = null;
	}

	function reset() {
		status = 'idle';
		apiKey = '';
		destinationWarning = null;
	}

	// Enter only submits from the idle input; it never confirms sending the key to
	// an untrusted backend.
	function enter(e: KeyboardEvent) {
		if (e.key === 'Enter' && status === 'idle') submit();
	}
</script>

<svelte:head>
	<link rel="icon" href="/images/logos/lightning-blue.svg" />
</svelte:head>

<div class="form-page">
	<div class="content">
		<div class="logo-icon">
			<Lightning />
		</div>
		<h2 class="title">Regenerate Link</h2>
		<p class="subtitle">Invalidate your current dashboard link and generate a new one.</p>

		{#if status === 'success'}
			<div class="status-msg success">
				Your dashboard link has been regenerated. Any previously shared links are now
				invalid.
			</div>
			<button class="form-btn" onclick={reset}>Done</button>
		{:else if status === 'error'}
			<div class="status-msg error">
				Something went wrong. Please check your API key and try again.
			</div>
			<button class="form-btn regen-btn" onclick={reset}>Try again</button>
		{:else if status === 'confirm'}
			<div class="status-msg warning">
				Your API key will be sent to <strong>{destinationWarning}</strong>, which is not the
				default backend. Only continue if this is your own server.
			</div>
			<div class="confirm-actions">
				<button class="form-btn cancel-btn" onclick={cancel}>Cancel</button>
				<button class="form-btn regen-btn" onclick={send}>Send &amp; regenerate</button>
			</div>
		{:else}
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
						<path
							d="M250 256.4Q413 180.4 550 556.4"
							marker-end="url(#arrowhead-regen)"
						/>
					</g>
					<defs>
						<marker
							markerWidth="6"
							markerHeight="6"
							refX="3"
							refY="3"
							viewBox="0 0 6 6"
							orient="auto"
							id="arrowhead-regen"
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
				disabled={status === 'loading'}
			/>
			<button class="form-btn regen-btn" onclick={submit} disabled={status === 'loading'}>
				{#if status === 'loading'}
					<div class="loader"></div>
				{:else}
					Regenerate
				{/if}
			</button>
		{/if}
	</div>
</div>

<style scoped>
	.content {
		box-shadow: 0 0 180px 2px var(--blue);
		max-width: calc(40ch + 6em);
	}
	.logo-icon {
		color: var(--blue);
	}
	.title {
		color: var(--blue);
	}
	.form-btn {
		background: var(--blue);
	}
	.regen-btn:hover:not(:disabled) {
		background: #1a8fd6;
	}
	.loader {
		border-color: rgba(255, 255, 255, 0.2);
		border-top-color: white;
	}
	.status-msg.success {
		background: #0d1f2b;
		color: var(--blue);
		border: 1px solid #1a3a4a;
	}
	.status-msg.warning {
		background: #2b1a0d;
		color: #e0a561;
		border: 1px solid #4a3320;
		text-align: left;
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
</style>
