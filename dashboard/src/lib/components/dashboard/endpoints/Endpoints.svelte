<script lang="ts">
	import EndpointList from './EndpointList.svelte';
	import EndpointFilter from './EndpointFilter.svelte';
	import { setParam, setParamNoReplace } from '$lib/params';
	import { type Endpoint, type EndpointFilterType, statusMatchesFilter } from '$lib/endpoints';

	type EndpointFreq = { [key: string]: { path: string; status: number; count: number } };

	function getFilteredEndpoints(freq: EndpointFreq, activeFilter: EndpointFilterType) {
		const endpoints: Endpoint[] = [];
		let maxCount = 0;
		for (const ep of Object.values(freq)) {
			if (statusMatchesFilter(ep.status, activeFilter)) {
				endpoints.push(ep);
				if (ep.count > maxCount) maxCount = ep.count;
			}
		}
		endpoints.sort((a, b) => b.count - a.count);
		return { endpoints: endpoints.slice(0, 50), maxCount };
	}

	function updateUrlParams(path: string | null, status: number | null): void {
		setParamNoReplace('path', path);
		setParam('status', status !== null ? status.toString() : null);
	}

	function handleEndpointSelection(path: string | null, status: number | null): void {
		if (path === null || status === null) {
			targetPath = null;
			targetStatus = null;
			updateUrlParams(null, null);
		} else if (targetPath === null) {
			targetPath = path;
			updateUrlParams(path, null);
		} else if (endpointData.endpoints.length > 1 && targetStatus === null) {
			targetStatus = status;
			updateUrlParams(targetPath, status);
		} else {
			targetPath = null;
			targetStatus = null;
			updateUrlParams(null, null);
		}
	}

	function handleFilterChange(value: EndpointFilterType): void {
		activeFilter = value;
	}

	let { endpointFreq, targetPath = $bindable<string | null>(null), targetStatus = $bindable<number | null>(null) }: {
		endpointFreq: EndpointFreq;
		targetPath: string | null;
		targetStatus: number | null;
	} = $props();

	let activeFilter = $state<EndpointFilterType>('all');
	const endpointData = $derived(getFilteredEndpoints(endpointFreq, activeFilter));
</script>

<div class="card">
	<div class="card-title header">
		Endpoints
		<div class="filter">
			<EndpointFilter {activeFilter} filterChange={handleFilterChange} />
		</div>
	</div>

	{#if endpointData.endpoints.length > 0}
		<EndpointList endpoints={endpointData.endpoints} maxCount={endpointData.maxCount} selectEndpoint={handleEndpointSelection} />
	{/if}
	{#if targetPath !== null}
		<div class="clear-row">
			<button class="clear-btn" onclick={() => handleEndpointSelection(null, null)}>Clear filter</button>
		</div>
	{/if}
</div>

<style>
	.card {
		min-height: 361px;
		display: flex;
		flex-direction: column;
	}

	.header {
		display: flex;
	}

	.filter {
		margin-left: auto;
	}

	.clear-row {
		margin-top: auto;
		text-align: right;
		padding: 0 20px 12px;
	}

	.clear-btn {
		font-size: 0.75em;
		color: var(--dim-text);
		cursor: pointer;
	}

	.clear-btn:hover {
		color: #ededed;
	}

	@media screen and (max-width: 470px) {
		.header {
			flex-wrap: wrap;
		}

		.filter {
			margin-left: 0;
			width: 100%;
		}
	}

	@media screen and (max-width: 1070px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 0 2em 0;
		}
	}
</style>
