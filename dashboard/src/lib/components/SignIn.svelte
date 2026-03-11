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

<div class="generate">
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
			API Key
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
		{#if loading}
			<div id="formBtn" class="grid place-items-center">
				<div class="loader"></div>
			</div>
		{:else}
			<button id="formBtn" onclick={submit}>{buttonLabels[type]}</button>
		{/if}
		<a href="/generate" class="generate-link">Don't have an API key? Generate one →</a>
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
	}
	.arrow {
		display: inline-block;
		width: 12px;
		height: 17px;
		margin-left: 3px;
		vertical-align: middle;
	}
	.input-error {
		outline: 1px solid var(--red) !important;
	}
	.error-msg {
		font-size: 0.8em;
		color: var(--red);
		text-align: left;
		margin-top: -1.8em;
		margin-bottom: 1.2em;
		padding: 0;
	}
	.loader {
		border: 3px solid #343434;
		border-top: 3px solid var(--highlight);
		width: 1em;
		height: 1em;
	}
	#formBtn {
		font-size: 0.9em;
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
