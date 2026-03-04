<script lang="ts">
	import { replaceState } from '$app/navigation';
	import { page } from '$app/state';
	import EndpointList from './EndpointList.svelte';
	import EndpointFilter from './EndpointFilter.svelte';
	import { type Endpoint, type EndpointFilterType } from '$lib/endpoints';

	type EndpointFreq = { [key: string]: { path: string; status: number; count: number } };

	function statusMatchesFilter(status: number, activeFilter: EndpointFilterType): boolean {
		return (
			activeFilter === 'all' ||
			(activeFilter === 'success' && status >= 200 && status <= 299) ||
			(activeFilter === 'redirect' && status >= 300 && status <= 399) ||
			(activeFilter === 'client' && status >= 400 && status <= 499) ||
			(activeFilter === 'server' && status >= 500)
		);
	}

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
		if (path === null) {
			page.url.searchParams.delete('path');
		} else {
			page.url.searchParams.set('path', path);
		}
		if (status === null) {
			page.url.searchParams.delete('status');
		} else {
			page.url.searchParams.set('status', status.toString());
		}
		replaceState(page.url, page.state);
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

	function clearSelection(): void {
		handleEndpointSelection(null, null);
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

	{#if targetPath !== null}
		<div class="mb-[-8px] flex px-[25px] pt-3 text-[13px] text-[#707070]">
			<div class="mr-3 flex-grow text-left">
				<div>
					{#if targetStatus === null}
						{targetPath}
					{:else}
						{targetPath}, status: {targetStatus}
					{/if}
				</div>
			</div>
			<button class="hover:text-[#ededed]" onclick={clearSelection}>Clear</button>
		</div>
	{/if}

	{#if endpointData.endpoints.length > 0}
		<EndpointList endpoints={endpointData.endpoints} maxCount={endpointData.maxCount} selectEndpoint={handleEndpointSelection} />
	{/if}
</div>

<style>
	.card {
		min-height: 361px;
	}

	.header {
		display: flex;
	}

	.filter {
		margin-left: auto;
	}

	@media screen and (max-width: 470px) {
		.header {
			flex-direction: column;
			gap: 8px;
		}

		.filter {
			margin-left: 0;
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
