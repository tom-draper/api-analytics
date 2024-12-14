<script lang="ts">
	import { onMount } from 'svelte';
	import { getServerURL } from '../lib/url';

	type State = 'sign-in' | 'loading';

	let state: State = 'sign-in';
	let apiKey = '';
	let queryString: string = '';

	async function submit() {
		if (!apiKey) return;

		setState('loading');

		try {
			const url = getServerURL();
			const response = await fetch(`${url}/api/user-id/${apiKey}`);

			if (response.status === 200) {
				const userID = await response.json();
				window.location.href = `/${page}/${userID.replaceAll('-', '')}${queryString ? `?${queryString}` : ''}`;
			} else {
				setState('sign-in');
			}
		} catch (e) {
			console.log(e);
			setState('sign-in');
		}
	}

	function enter(e: KeyboardEvent) {
		if (e.keyCode === 13) {
			submit();
		}
	}

	function setState(value: State) {
		state = value;
	}

	onMount(() => {
		const params = new URLSearchParams(window.location.search);
		queryString = params.toString();
		console.log(queryString);
	});

	export let page: 'dashboard' | 'monitoring';
</script>

<div class="generate">
	<div class="content">
		{#if page === 'dashboard'}
			<h2>Dashboard</h2>
		{:else if page === 'monitoring'}
			<h2>Monitoring</h2>
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
			class:no-display={state !== 'sign-in'}>Load</button
		>
		<button id="formBtn" class:no-display={state !== 'loading'}>
			<div class="spinner">
				<div class="loader" />
			</div>
		</button>
	</div>
	<div class="details">
		<div class="keep-secure">Keep your API key safe and secure.</div>
		<div class="highlight logo">API Analytics</div>
		<img class="footer-logo" src="img/logos/lightning-green.png" alt="" />
	</div>
</div>

<style scoped>
	.spinner {
		height: auto;
	}
	.loader {
		border: 3px solid #343434;
		border-top: 3px solid var(--highlight);
		height: 10px;
		width: 10px;
	}
</style>
