<script lang="ts">
	import { getServerURL } from '$lib/url';
	import { formatPath } from '$lib/path';
	import { page } from '$app/state';

	let apiKey: string = '';
	let loading: boolean = false;

	let params: string;
	$: params = page.url.searchParams.toString();

	async function submit() {
		if (!apiKey) {
			return;
		}

		loading = true;

		const url = getServerURL();

		try {
			const response = await fetch(`${url}/api/user-id/${apiKey}`);

			if (response.status === 200) {
				const userID = await response.json();
				window.location.href = formatPath(`/${page}/${userID.replaceAll('-', '')}`, params);
			}
		} catch (e) {
			console.log(e);
		}

		loading = false;
	}

	function enter(e: KeyboardEvent) {
		if (e.keyCode === 13) {
			submit();
		}
	}

	export let type: 'dashboard' | 'monitor' | 'explorer';
</script>

<div class="generate">
	<div class="content place-items-center">
		{#if type === 'dashboard'}
			<h2 class="font-bold">Dashboard</h2>
		{:else if type === 'monitor'}
			<h2 class="font-bold">Monitor</h2>
		{:else if type === 'explorer'}
			<h2 class="font-bold">Explorer</h2>
		{/if}
		<input type="text" bind:value={apiKey} placeholder="Enter API key" on:keydown={enter} />
		<button id="formBtn" on:click={submit} class="text-sm" class:no-display={loading}>Load</button>
		<div id="formBtn" class="grid place-items-center" class:no-display={!loading}>
			<div class="h-auto place-items-center">
				<div class="loader !h-[1em] !w-[1em]"></div>
			</div>
		</div>
	</div>
</div>

<style scoped>
	h2 {
		font-size: 2em;
		color: var(--highlight);
		margin-bottom: 1em;
	}
	.loader {
		border: 3px solid #343434;
		border-top: 3px solid var(--highlight);
	}
</style>
