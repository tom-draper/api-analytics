<script lang="ts">
	import { type Filter } from '$lib/filter';
	import type { ComponentType } from 'svelte';

	let hidden: boolean = false;

	function toggleHidden() {
		hidden = !hidden;
	}

	export let title: string, content: ComponentType, filter: Filter, data: DashboardData;
</script>

<div class="text-left text-[16px] text-[var(--faint-text)]">
	<button
		onclick={toggleHidden}
		class="m-auto flex w-full px-2 py-2 text-[var(--faint-text)] hover:text-[#ededed] rounded"
		class:!text-[#ededed]={!hidden}
	>
		<div class="mr-auto">
			{title}
		</div>
		<div class="my-auto">
			{#if hidden}
				<svg
					xmlns="http://www.w3.org/2000/svg"
					fill="none"
					viewBox="0 0 24 24"
					stroke-width="1.5"
					stroke="currentColor"
					class="h-[1em] w-[1em]"
				>
					<path stroke-linecap="round" stroke-linejoin="round" d="m4.5 15.75 7.5-7.5 7.5 7.5" />
				</svg>
			{:else}
				<svg
					xmlns="http://www.w3.org/2000/svg"
					fill="none"
					viewBox="0 0 24 24"
					stroke-width="2"
					stroke="currentColor"
					class="h-[1em] w-[1em]"
				>
					<path stroke-linecap="round" stroke-linejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
				</svg>
			{/if}
		</div>
	</button>

	<div class="pb-2 pt-1" class:no-display={hidden || !data}>
		<div class="rounded border border-solid border-[#2e2e2e] text-[14px]">
			{#if content}
				<svelte:component this={content} bind:filter={filter} bind:data={data} />
			{/if}
		</div>
	</div>
</div>
