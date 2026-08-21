<script lang="ts">
	import { getServerURL } from '$lib/url';
	import { formatPath } from '$lib/path';
	import { page } from '$app/state';
	import Lightning from '$components/Lightning.svelte';

	type State = 'idle' | 'loading' | 'error';

	let status = $state<State>('idle');
	let apiKey = $state('');

	async function submit() {
		if (!apiKey) return;

		status = 'loading';

		try {
			const url = getServerURL();
			// Sent as a POST body (no custom header) to keep the API key out of
			// URLs and to stay a CORS simple request, matching the delete page.
			const response = await fetch(`${url}/api/regenerate-user-id`, {
				method: 'POST',
				body: JSON.stringify({ api_key: apiKey })
			});

			if (response.status === 200) {
				const userID = await response.json();
				const params = page.url.searchParams.toString();
				window.location.href = formatPath(`/dashboard/${userID.replaceAll('-', '')}`, params);
				return;
			}

			status = 'error';
		} catch (e) {
			console.log(e);
			status = 'error';
		}
	}

	function retry() {
		status = 'idle';
	}

	function enter(e: KeyboardEvent) {
		if (e.key === 'Enter') submit();
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

		{#if status === 'error'}
			<div class="status-msg error">
				Something went wrong. Please check your API key and try again.
			</div>
			<button class="form-btn regen-btn" onclick={retry}>Try again</button>
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
</style>
