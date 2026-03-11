<script lang="ts">
	import { getServerURL } from '$lib/url';

	type State = 'idle' | 'loading' | 'deleted' | 'error';

	let state: State = $state('idle');
	let apiKey = $state('');

	async function submit() {
		if (!apiKey) return;

		state = 'loading';

		try {
			const url = getServerURL();
			const response = await fetch(`${url}/api/delete/${apiKey}`);
			state = response.status === 200 ? 'deleted' : 'error';
		} catch (e) {
			console.log(e);
			state = 'error';
		}
	}

	function enter(e: KeyboardEvent) {
		if (e.key === 'Enter') submit();
	}
</script>

<div class="generate">
	<div class="content">
		<h2 class="title">Delete Account</h2>
		<p class="subtitle">Permanently delete all data associated with your API key.</p>
		{#if state === 'deleted'}
			<div class="status-msg success">Your account and all associated data has been deleted.</div>
		{:else if state === 'error'}
			<div class="status-msg error">Something went wrong. Please check your API key and try again.</div>
		{:else}
			<input
				type="text"
				bind:value={apiKey}
				placeholder="Enter API key"
				onkeydown={enter}
				disabled={state === 'loading'}
			/>
			{#if state === 'loading'}
				<div id="formBtn" class="grid place-items-center">
					<div class="loader"></div>
				</div>
			{:else}
				<button id="formBtn" class="delete-btn" onclick={submit}>Delete</button>
			{/if}
		{/if}
	</div>
</div>

<style scoped>
	.title {
		font-size: 2em;
		font-weight: 700;
		color: var(--red);
		margin-bottom: 0.25em;
	}
	.subtitle {
		font-size: 0.9em;
		color: var(--dim-text);
		margin-bottom: 1.8em;
		padding: 0;
	}
	.delete-btn {
		background: var(--red) !important;
	}
	.delete-btn:hover {
		background: #c94f4f !important;
	}
	.status-msg {
		font-size: 0.9em;
		padding: 0.8em 1em;
		border-radius: 4px;
		text-align: center;
	}
	.status-msg.success {
		background: #0d2b1a;
		color: var(--highlight);
		border: 1px solid #1a4a2e;
	}
	.status-msg.error {
		background: #2b0d0d;
		color: var(--red);
		border: 1px solid #4a1a1a;
	}
	.loader {
		border: 3px solid #343434;
		border-top: 3px solid var(--red);
		width: 1em;
		height: 1em;
	}
</style>
