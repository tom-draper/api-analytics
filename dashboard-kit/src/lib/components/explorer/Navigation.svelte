<script lang="ts">
	import Lightning from '$components/Lightning.svelte';
	import Expandable from './Expandable.svelte';
</script>

<nav
	class="fixed flex h-full w-[20em] flex-col border-r border-[#2e2e2e] bg-[var(--light-background)]"
>
	<div class="flex-grow p-4">
		<div class="flex pb-4">
			<h2 class="text-left text-[var(--faded-text)]">Filters</h2>
			<div class="flex text-[var(--faint-text)] items-center ml-auto text-xs">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					fill="none"
					viewBox="0 0 24 24"
					stroke-width="1.5"
					stroke="currentColor"
					class="size-4"
				>
					<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
				</svg>

				<div>Reset</div>
			</div>
		</div>
		<Expandable title="Timespan" />
		<Expandable title="Status" />
		<Expandable title="Test" />
	</div>

	<div class="grid h-[6em] place-items-center">
		<div class="h-[21px] text-[var(--highlight)]">
			<Lightning />
		</div>
	</div>
</nav>

<style>
</style>
