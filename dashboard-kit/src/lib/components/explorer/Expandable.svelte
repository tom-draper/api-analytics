<script lang="ts">
	let hidden: boolean = true;

    function toggleHidden() {
        hidden = !hidden
    }

	export let title: string;
</script>

<div class="text-[var(--faint-text)] text-sm text-left">
	<button onclick={toggleHidden} class="flex m-auto py-2 w-full">
		<div class="mr-auto">
			{title}
		</div>
		<div class="mt-auto">
			<svg
				xmlns="http://www.w3.org/2000/svg"
				fill="none"
				viewBox="0 0 24 24"
				stroke-width="2"
				stroke="currentColor"
				class="h-4 w-4 text-[var(--faint-text)]"
			>
				<path stroke-linecap="round" stroke-linejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
			</svg>
		</div>
	</button>

	<div class="pb-2 pt-1" class:no-display={hidden}>
		<div class="p-2 border border-solid border-[#2e2e2e] rounded">
			content
		</div>
    </div>
</div>
