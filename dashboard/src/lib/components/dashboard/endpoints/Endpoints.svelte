<script lang="ts">
	import { replaceState } from '$app/navigation';
	import { page } from '$app/state';
	import EndpointList from './EndpointList.svelte';
	import EndpointFilter from './EndpointFilter.svelte';
	import { type Endpoint, type EndpointFilterType, getEndpoints } from '$lib/endpoints';

	let activeFilter: EndpointFilterType = 'all';
	let endpoints: Endpoint[] = [];
	let maxCount = 0;

	// Derived state - recalculate endpoints when data or filter changes
	$: if (data && activeFilter) {
		({ endpoints, maxCount } = getEndpoints(data, activeFilter, ignoreParams));
	}

	// URL Parameter Management
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

	function handleEndpointSelection(
		e: CustomEvent<{ path: string | null; status: number | null }>
	): void {
		const { path, status } = e.detail;

		if (path === null || status === null) {
			// Reset selection
			targetPath = null;
			targetStatus = null;
			updateUrlParams(null, null);
		} else if (targetPath === null) {
			// Set path first
			targetPath = path;
			updateUrlParams(path, null);
		} else if (endpoints.length > 1 && targetStatus === null) {
			// Path already set, now set status if multiple endpoints exist
			targetStatus = status;
			updateUrlParams(targetPath, status);
		} else {
			// Reset if both already set
			targetPath = null;
			targetStatus = null;
			updateUrlParams(null, null);
		}
	}

	function handleFilterChange(e: CustomEvent<EndpointFilterType>): void {
		activeFilter = e.detail;
	}

	function clearSelection(): void {
		handleEndpointSelection(
			new CustomEvent('selectEndpoint', {
				detail: { path: null, status: null }
			})
		);
	}

	export let data: RequestsData,
		targetPath: string | null,
		targetStatus: number | null,
		ignoreParams: boolean;
</script>

<div class="card">
	<div class="card-title flex">
		Endpoints
		<div class="ml-auto">
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
			<button class="hover:text-[#ededed]" on:click={clearSelection}>Clear</button>
		</div>
	{/if}

	{#if endpoints.length > 0}
		<EndpointList {endpoints} {maxCount} selectEndpoint={handleEndpointSelection} />
	{/if}
</div>

<style>
	.card {
		min-height: 361px;
	}

	@media screen and (max-width: 1030px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 0 2em 0;
		}
	}
</style>
