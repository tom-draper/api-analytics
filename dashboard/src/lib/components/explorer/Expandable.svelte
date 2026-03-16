<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { Filter } from '$lib/filter';

	let { title, filter, children }: {
		title: string;
		filter: Filter | undefined;
		children: Snippet;
	} = $props();
	let hidden = $state(false);
</script>

<div class="text-left text-[16px] text-[var(--faint-text)]">
	<button
		onclick={() => (hidden = !hidden)}
		class="m-auto flex w-full rounded px-2 py-2 text-[var(--faint-text)] hover:text-[var(--faded-text)]"
		class:!text-[var(--faded-text)]={!hidden}
	>
		<div class="mr-auto">{title}</div>
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

	<div class="pb-2 pt-1" class:no-display={hidden || !filter}>
		<div class="rounded border border-solid border-[var(--border)] text-[14px]">
			{@render children()}
		</div>
	</div>
</div>
