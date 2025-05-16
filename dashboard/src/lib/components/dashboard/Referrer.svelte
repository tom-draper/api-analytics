<script lang="ts">
	import { replaceState } from '$app/navigation';
	import { page } from '$app/state';
	import { ColumnIndex } from '$lib/consts';
	import { setParam } from '$lib/params';

	type ReferrerFreq = Map<string, { referrer: string; count: number }>;

	type MapValue<A> = A extends Map<unknown, infer V> ? V : never;

	function endpointFreq(data: RequestsData) {
		const freq: ReferrerFreq = new Map();
		for (const row of data) {
			// Create groups of endpoints by path + status
			const referrer = ignoreParams ? row[ColumnIndex.Path].split('?')[0] : row[ColumnIndex.Path];

			let referrerCount = freq.get(referrer);
			if (!referrerCount) {
				referrerCount = {
					count: 0,
					referrer
				};
				freq.set(referrer, referrerCount);
			}
			referrerCount.count++;
		}

		return freq;
	}

	function removeEndpointParams() {
		page.url.searchParams.delete('path');
		replaceState(page.url, page.state);
	}

	function setReferrerParam(referrer: string | null) {
		setParam('referrer', referrer);
	}

	function setTargetEndpoint(referrer: string | null) {
		if (referrer === null) {
			// Trigger reset if input is null
			targetReferrer = null;
			removeEndpointParams();
		} else {
			targetReferrer = referrer;
			setReferrerParam(referrer);
		}
	}

	function getReferrers(data: RequestsData) {
		const freq = endpointFreq(data);

		// Convert object to list
		const freqArr: MapValue<ReferrerFreq>[] = [];
		let maxCount = 0;
		for (const value of freq.values()) {
			freqArr.push(value);
			if (value.count > maxCount) {
				maxCount = value.count;
			}
		}

		freqArr.sort((a, b) => {
			return b.count - a.count;
		});

		return {
			referrers: freqArr.slice(0, 50),
			maxCount
		};
	}

	let referrers: {
		referrer: string;
		count: number;
	}[];
	let maxCount: number;

	$: if (data) {
		({ referrers, maxCount } = getReferrers(data));
	}

	export let data: RequestsData, targetReferrer: string | null, ignoreParams: boolean;
</script>

<div class="card">
	<div class="card-title">Referrer</div>

	{#if referrers != undefined}
		<div class="endpoints">
			{#each referrers as referrer, i}
				<div class="endpoint-container">
					<button
						class="endpoint"
						id="endpoint-{i}"
						on:click={() => setTargetEndpoint(referrer.referrer)}
					>
						<div class="path">
							<span class="font-semibold">{referrer.count.toLocaleString()}</span>
							{referrer.referrer}
						</div>
						<div class="background" style="width: {(referrer.count / maxCount) * 100}%"></div>
					</button>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style scoped>
	.card {
		min-height: 361px;
		margin-top: 0;
		margin-left: 2em;
	}
	.card-title {
		display: flex;
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
		background: linear-gradient(270deg, transparent, var(--background));
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
		color: var(--light-background);

		background: var(--highlight);
		text-align: left;
		position: relative;
		font-size: 0.85em;
		height: 100%;
		position: absolute;
		top: 0;
	}
	@media screen and (max-width: 1030px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 0 2em 0;
		}
	}
</style>
