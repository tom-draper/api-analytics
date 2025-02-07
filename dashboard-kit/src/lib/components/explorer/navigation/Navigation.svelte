<script lang="ts">
	import Lightning from '$components/Lightning.svelte';
	import Expandable from '../Expandable.svelte';
	import Status from './filters/Status.svelte';
</script>

<nav
	class="fixed flex h-full w-[20em] flex-col border-r border-[#2e2e2e] bg-[var(--light-background)]"
>
	<div class="flex-grow p-2">
		<div class="flex px-2 pb-4 pt-2">
			<h2 class="text-left text-[var(--faded-text)]">Filters</h2>
			<button
				class="ml-auto flex items-center rounded border border-[#2e2e2e] px-2 text-xs text-[var(--faint-text)] hover:text-[#ededed]"
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
		</div>
		<Expandable title="Timespan" content={Status}/>
		<Expandable title="Status" content={Status} />
		<Expandable title="Test" content={Status}/>
	</div>

	<div class="grid h-[6em] place-items-center">
		<div class="h-[21px] text-[var(--highlight)]">
			<Lightning />
		</div>
	</div>
</nav>

<style>
</style>
