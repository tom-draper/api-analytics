<script lang="ts">
    $: color;
	export let checked: boolean,
		label: string,
		color: string;
</script>

<div class="flex items-center gap-2 px-2 py-2">
	<button class="flex items-center gap-2 hover:text-[#ededed] px-1" onclick={() => (checked = !checked)}>
		<div class="flex items-center gap-2">
			<div class="grid h-4 w-4 place-items-center rounded-[3px] border border-solid border-[#2e2e2e]" style={`${checked ? `background: ${color}` : ''}`}>
				{#if checked}
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="4"
						stroke="currentColor"
						class="size-3 text-[var(--light-background)]"
					>
						<path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" />
					</svg>
				{/if}
			</div>
			<div>
				{label}
			</div>
		</div>
	</button>
</div>
