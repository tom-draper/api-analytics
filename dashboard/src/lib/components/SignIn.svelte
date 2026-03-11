<script lang="ts">
	import { getServerURL } from '$lib/url';
	import { formatPath } from '$lib/path';
	import { page } from '$app/state';

	let { type }: { type: 'dashboard' | 'monitor' | 'explorer' } = $props();

	let apiKey = $state('');
	let loading = $state(false);
	let error = $state(false);

	const params = $derived(page.url.searchParams.toString());

	const subtitles: Record<typeof type, string> = {
		dashboard: 'View your API analytics',
		monitor: 'Monitor your API uptime',
		explorer: 'Explore your request data',
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
		<h2 class="title">
			{#if type === 'dashboard'}Dashboard
			{:else if type === 'monitor'}Monitor
			{:else}Explorer{/if}
		</h2>
		<p class="subtitle">{subtitles[type]}</p>
		<input
			type="text"
			bind:value={apiKey}
			placeholder="Enter API key"
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
			<button id="formBtn" onclick={submit}>Load</button>
		{/if}
	</div>
</div>

<style scoped>
	.title {
		font-size: 2em;
		font-weight: 700;
		color: var(--highlight);
		margin-bottom: 0.25em;
	}
	.subtitle {
		font-size: 0.9em;
		color: var(--dim-text);
		margin-bottom: 1.8em;
		padding: 0;
	}
	.input-error {
		outline: 1px solid var(--red) !important;
	}
	.error-msg {
		font-size: 0.8em;
		color: var(--red);
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
</style>
