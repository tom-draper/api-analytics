<script lang="ts">
	import { getServerURL } from '$lib/url';
	import Lightning from '$components/Lightning.svelte';

	type State = 'idle' | 'loading' | 'success' | 'error';

	let state = $state<State>('idle');
	let apiKey = $state('');

	async function submit() {
		if (!apiKey) return;

		state = 'loading';

		try {
			const url = getServerURL();
			const response = await fetch(`${url}/api/regenerate-user-id`, {
				method: 'POST',
				headers: { 'X-AUTH-TOKEN': apiKey },
			});
			state = response.status === 200 ? 'success' : 'error';
		} catch (e) {
			console.log(e);
			state = 'error';
		}
	}

	function reset() {
		state = 'idle';
		apiKey = '';
	}

	function enter(e: KeyboardEvent) {
		if (e.key === 'Enter') submit();
	}
</script>

<svelte:head>
	<link rel="icon" href="/images/logos/lightning-blue.svg" />
</svelte:head>

<div class="generate">
	<div class="content">
		<div class="logo-icon">
			<Lightning />
		</div>
		<h2 class="title">Regenerate Link</h2>
		<p class="subtitle">Invalidate your current dashboard link and generate a new one.</p>

		{#if state === 'success'}
			<div class="status-msg success">
				Your dashboard link has been regenerated. Any previously shared links are now invalid.
			</div>
			<button class="form-btn" onclick={reset}>Done</button>
		{:else if state === 'error'}
			<div class="status-msg error">
				Something went wrong. Please check your API key and try again.
			</div>
			<button class="form-btn regen-btn" onclick={reset}>Try again</button>
		{:else}
			<label class="input-label" for="api-key">
				Enter API Key
				<svg class="arrow" viewBox="240 170 320 400" fill="none" xmlns="http://www.w3.org/2000/svg">
					<g stroke-width="31" stroke="currentColor" stroke-linecap="square" transform="matrix(1,0,0,1,-4,0)">
						<path d="M250 256.4Q413 180.4 550 556.4" marker-end="url(#arrowhead-regen)"/>
					</g>
					<defs>
						<marker markerWidth="6" markerHeight="6" refX="3" refY="3" viewBox="0 0 6 6" orient="auto" id="arrowhead-regen">
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
				disabled={state === 'loading'}
			/>
			<button class="form-btn regen-btn" onclick={submit} disabled={state === 'loading'}>
				{#if state === 'loading'}
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
		height: 36px;
		margin: 0 auto 1em;
	}
	.title {
		font-size: 2em;
		font-weight: 700;
		color: var(--blue);
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
	input::placeholder {
		color: #707070;
	}
	.form-btn {
		font-size: 0.9em;
		height: 40px;
		border-radius: 4px;
		padding: 0 20px;
		border: none;
		cursor: pointer;
		width: auto;
		background: var(--blue);
		font-family: 'Noto Sans', 'Geist' !important;
		font-weight: 400;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}
	.regen-btn:hover:not(:disabled) {
		background: #1a8fd6;
	}
	.loader {
		border: 3px solid rgba(255, 255, 255, 0.2);
		border-top: 3px solid white;
		width: 1em;
		height: 1em;
	}
	.status-msg {
		font-size: 0.9em;
		padding: 0.9em 1.2em;
		border-radius: 4px;
		text-align: left;
		line-height: 1.5;
		margin-bottom: 1.2em;
	}
	.status-msg.success {
		background: #0d1f2b;
		color: var(--blue);
		border: 1px solid #1a3a4a;
	}
	.status-msg.error {
		background: #2b0d0d;
		color: var(--red);
		border: 1px solid #4a1a1a;
	}
</style>
