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
				window.location.href = formatPath(`/${type}/${userID.replaceAll('-', '')}`, params);
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
		<h2 class="font-bold text-[2em] mb-[1em] text-[var(--highlight)]">
			{#if type === 'dashboard'}
				Dashboard
			{:else if type === 'monitor'}
				Monitor
			{:else if type === 'explorer'}
				Explorer
			{/if}
		</h2>
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
	.loader {
		border: 3px solid #343434;
		border-top: 3px solid var(--highlight);
	}
</style>
