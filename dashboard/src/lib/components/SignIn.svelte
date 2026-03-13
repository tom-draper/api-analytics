<script lang="ts">
	import { getServerURL } from '$lib/url';
	import { formatPath } from '$lib/path';
	import { page } from '$app/state';
	import Lightning from '$components/Lightning.svelte';

	let { type }: { type: 'dashboard' | 'monitor' | 'explorer' } = $props();

	let apiKey = $state('');
	let loading = $state(false);
	let error = $state(false);

	const params = $derived(page.url.searchParams.toString());

	const subtitles: Record<typeof type, string> = {
		dashboard: 'View your API analytics',
		monitor: 'Monitor your API uptime',
		explorer: 'Explore your logged requests',
	};

	const buttonLabels: Record<typeof type, string> = {
		dashboard: 'View',
		monitor: 'View',
		explorer: 'Explore',
	};

	async function submit() {
		if (!apiKey) return;

		loading = true;
		error = false;

		const url = getServerURL();

		try {
			const response = await fetch(`${url}/api/user-id/${apiKey}`);

			if (response.status === 200) {
				const userID = await response.json();
				window.location.href = formatPath(`/${type}/${userID.replaceAll('-', '')}`, params);
				return;
			}
		} catch (e) {
			console.log(e);
		}

		error = true;
		loading = false;
	}

	function enter(e: KeyboardEvent) {
		if (e.key === 'Enter') submit();
	}
</script>

<div class="form-page">
	<div class="content">
		<div class="logo-icon">
			<Lightning />
		</div>
		<h2 class="title">
			{#if type === 'dashboard'}Dashboard
			{:else if type === 'monitor'}Monitor
			{:else}Explorer{/if}
		</h2>
		<p class="subtitle">{subtitles[type]}</p>
		<label class="input-label" for="api-key">
			Enter API Key
			<svg class="arrow" viewBox="240 170 320 400" fill="none" xmlns="http://www.w3.org/2000/svg">
				<g stroke-width="31" stroke="currentColor" stroke-linecap="square" transform="matrix(1,0,0,1,-4,0)">
					<path d="M250 256.4Q413 180.4 550 556.4" marker-end="url(#arrowhead)"/>
				</g>
				<defs>
					<marker markerWidth="6" markerHeight="6" refX="3" refY="3" viewBox="0 0 6 6" orient="auto" id="arrowhead">
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
			class:input-error={error}
		/>
		{#if error}
			<p class="error-msg">API key not found. Please check and try again.</p>
		{/if}
		<button class="form-btn" onclick={submit} disabled={loading}>
			{#if loading}
				<div class="loader"></div>
			{:else}
				{buttonLabels[type]}
			{/if}
		</button>
		<a href="/sign-up" class="generate-link">Don't have an API key? Generate one →</a>
	</div>
</div>

<style scoped>
	.logo-icon {
		color: var(--highlight);
	}
	.title {
		color: var(--highlight);
	}
	.input-error {
		outline: 1px solid var(--red) !important;
	}
	.error-msg {
		font-size: 0.8em;
		color: var(--red);
		text-align: center;
		margin-top: -1.8em;
		margin-bottom: 1.2em;
		padding: 0;
	}
	.form-btn {
		background: var(--highlight);
		width: 100px;
	}
	.generate-link {
		display: block;
		margin-top: 1.5em;
		font-size: 0.8em;
		color: var(--dim-text);
		font-weight: normal;
		text-decoration: none;
		transition: color 0.15s;
	}
	.generate-link:hover {
		color: var(--highlight);
	}
</style>
