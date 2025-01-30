<script lang="ts">
	import { onMount } from 'svelte';
	import { getServerURL } from '$lib/url';
	import { formatPath } from '$lib/path';

	let apiKey: string = '';
	let queryString: string = '';
	let loading: boolean = false;

	async function submit() {
		if (!apiKey) {
            return;
        }

		loading = true;
		
		try {
			const url = getServerURL();
			const response = await fetch(`${url}/api/user-id/${apiKey}`);

			if (response.status === 200) {
				const userID = await response.json();
				window.location.href = formatPath(`/${page}/${userID.replaceAll('-', '')}`, queryString);
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

	onMount(() => {
		const params = new URLSearchParams(window.location.search);
		queryString = params.toString();
	});

	export let page: 'dashboard' | 'monitor' | 'explorer';
</script>

<div class="generate">
	<div class="content place-items-center">
		{#if page === 'dashboard'}
			<h2 class="font-bold">Dashboard</h2>
		{:else if page === 'monitor'}
			<h2 class="font-bold">Monitor</h2>
		{:else if page === 'explorer'}
			<h2 class="font-bold">Explorer</h2>
		{/if}
		<input
			type="text"
			bind:value={apiKey}
			placeholder="Enter API key"
			on:keydown={enter}
		/>
		<button
			id="formBtn"
			on:click={submit}
			class="text-sm"
			class:no-display={loading}>Load</button
		>
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
