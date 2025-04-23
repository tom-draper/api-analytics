<script lang="ts">
	import Lightning from '$components/Lightning.svelte';
	import { defaultFilter, type Filter } from '$lib/filter';
	import Expandable from '../Expandable.svelte';
	import Status from './filters/Status.svelte';
	import Timespan from './filters/Timespan.svelte';
	import Method from './filters/Method.svelte';
	import Hostname from './filters/Hostname.svelte';
	import ResponseTime from './filters/ResponseTime.svelte';

	function resetFilter() {
		filter = defaultFilter(data.requests);
	}

	export let data: DashboardData, filteredRequests: RequestsData, filter: Filter;
</script>

<nav
	class="fixed flex h-full w-[20em] flex-col border-r border-[#2e2e2e] bg-[var(--light-background)] overflow-y-auto"
>
	<div class="flex-grow p-2">
		<div class="flex px-2 pb-4 pt-2">
			<h2 class="text-left text-[var(--faded-text)]">Filters</h2>
			{#if data && filteredRequests && data.requests.length !== filteredRequests.length}
				<button
					class="ml-auto flex items-center rounded border border-[#2e2e2e] px-2 text-xs text-[var(--faint-text)] hover:text-[#ededed]"
					onclick={resetFilter}
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="mr-1 size-3"
					>
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
					</svg>

					<div>Reset</div>
				</button>
			{/if}
		</div>
		<Expandable title="Timespan" content={Timespan} bind:filter bind:data />
		<Expandable title="Status" content={Status} bind:filter bind:data />
		<Expandable title="Method" content={Method} bind:filter bind:data />
		<Expandable title="Hostname" content={Hostname} bind:filter bind:data />
		<!-- <Expandable title="Hostname" content={Hostname} bind:filter bind:data /> -->
		<!-- <Expandable title="Hostname" content={Hostname} bind:filter bind:data /> -->
		<!-- <Expandable title="Hostname" content={Hostname} bind:filter bind:data /> -->
		<Expandable title="Response Time" content={ResponseTime} bind:filter bind:data />
	</div>

	<div class="grid place-items-center my-10">
		<div class="h-[28px] text-[var(--highlight)]">
			<Lightning />
		</div>
	</div>
</nav>
