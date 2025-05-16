<script lang="ts">
	import { getServerURL } from '$lib/url';

	type State = 'delete' | 'loading' | 'deleted' | 'error';

	let state: State = 'delete';
	let apiKey = '';
	async function submit() {
		setState('loading');

		try {
			const url = getServerURL();
			const response = await fetch(`${url}/api/delete/${apiKey}`);

			if (response.status === 200) {
				setState('deleted');
			} else {
				setState('error');
			}
		} catch (e) {
			console.log(e);
			setState('error');
		}
	}

	function enter(e) {
		if (e.keyCode === 13) {
			submit();
		}
	}

	function setState(value: State) {
		state = value;
	}
</script>

<div class="generate">
	<div class="content">
		<h2 class="font-bold">Delete Account</h2>
		<input type="text" bind:value={apiKey} placeholder="Enter API key" on:keydown={enter} />
		<button id="formBtn" on:click={submit} class="text-sm" class:no-display={state != 'delete'}
			>Delete</button
		>
		<button id="formBtn" class="text-sm" class:no-display={state != 'loading'}>
			<div class="spinner">
				<div class="loader" />
			</div>
		</button>
		<button id="formBtn" class="copied-btn text-sm" class:no-display={state != 'deleted'}
			>Deleted</button
		>
		<button id="formBtn" class="text-sm" class:no-display={state != 'error'}>Error</button>
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
