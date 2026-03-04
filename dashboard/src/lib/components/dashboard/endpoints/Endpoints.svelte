<script lang="ts">
	import { replaceState } from '$app/navigation';
	import { page } from '$app/state';
	import EndpointList from './EndpointList.svelte';
	import EndpointFilter from './EndpointFilter.svelte';
	import { type Endpoint, type EndpointFilterType, getEndpoints } from '$lib/endpoints';

	let activeFilter = $state<EndpointFilterType>('all');
	const endpointData = $derived(
		data ? getEndpoints(data, activeFilter, ignoreParams) : { endpoints: [], maxCount: 0 }
	);

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

	function handleEndpointSelection(path: string | null, status: number | null): void {
		if (path === null || status === null) {
			// Reset selection
			targetPath = null;
			targetStatus = null;
			updateUrlParams(null, null);
		} else if (targetPath === null) {
			// Set path first
			targetPath = path;
			updateUrlParams(path, null);
		} else if (endpointData.endpoints.length > 1 && targetStatus === null) {
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

	function handleFilterChange(value: EndpointFilterType): void {
		activeFilter = value;
	}

	function clearSelection(): void {
		handleEndpointSelection(null, null);
	}

	let { data, targetPath = $bindable<string | null>(null), targetStatus = $bindable<number | null>(null), ignoreParams = $bindable(false) }: { data: RequestsData; targetPath: string | null; targetStatus: number | null; ignoreParams: boolean } = $props();
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

	@media screen and (max-width: 1030px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 0 2em 0;
		}
	}
</style>
