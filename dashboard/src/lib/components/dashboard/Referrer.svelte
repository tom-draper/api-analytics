<script lang="ts">
	import { replaceState } from '$app/navigation';
	import { page } from '$app/state';
	import { setParam } from '$lib/params';
	import type { ReferrerBar } from '$lib/aggregate';

	function setTargetReferrer(referrer: string) {
		if (targetReferrer === referrer) {
			targetReferrer = null;
			page.url.searchParams.delete('referrer');
			replaceState(page.url, page.state);
		} else {
			targetReferrer = referrer;
			setParam('referrer', referrer);
		}
	}

	let { referrerBars, targetReferrer = $bindable<string | null>(null) }: {
		referrerBars: ReferrerBar[];
		targetReferrer: string | null;
	} = $props();
</script>

<div class="card">
	<div class="card-title">Referrer</div>
	<div class="endpoints">
		{#each referrerBars as bar, i}
			<div class="endpoint-container">
				<button
					class="endpoint"
					id="endpoint-{i}"
					class:selected={targetReferrer === bar.referrer}
					onclick={() => setTargetReferrer(bar.referrer)}
				>
					<div class="path">
						<span class="font-semibold">{bar.count.toLocaleString()}</span>
						{bar.referrer}
					</div>
					<div class="background" style="width: {bar.height * 100}%"></div>
				</button>
			</div>
		{/each}
	</div>
</div>

<style scoped>
	.card {
		min-height: 361px;
		margin-top: 0;
		margin-left: 2em;
	}
	.endpoints {
		margin: 0.9em 20px 0.6em;
	}
	.endpoint {
		border-radius: 3px;
		margin: 5px 0;
		color: var(--light-background);
		text-align: left;
		position: relative;
		font-size: 0.85em;
		width: 100%;
		cursor: pointer;
	}
	.endpoint:hover {
		background: linear-gradient(270deg, transparent, #444);
	}
	.selected {
		background: linear-gradient(270deg, transparent, #444);
	}
	.path {
		position: relative;
		flex-grow: 1;
		z-index: 1;
		pointer-events: none;
		color: #505050;
		padding: 3px 12px;
		overflow-wrap: break-word;
		font-family: 'Noto Sans' !important;
	}
	.endpoint-container {
		display: flex;
	}
	.background {
		border-radius: 3px;
		background: var(--highlight);
		text-align: left;
		position: absolute;
		top: 0;
		height: 100%;
		font-size: 0.85em;
	}
	@media screen and (max-width: 1600px) {
		.card {
			margin-left: 0;
			width: 100%;
			min-height: unset;
		}
	}
	@media screen and (max-width: 1030px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 0 2em 0;
		}
	}
</style>
