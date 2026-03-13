<script lang="ts">
	export let data;

	let showing = data.faq.map(() => false);

	function toggleAnswer(index: number) {
		showing[index] = !showing[index];
	}
</script>

<div class="info-page-container">
	<h1>Frequently Asked Questions</h1>

	{#each data.faq as item, i}
		<div class="mb-4">
			<button class="question-btn" on:click={() => toggleAnswer(i)}>
				<div class="m-auto flex-1">
					{item.question}
				</div>
				<div class="dropdown-icon-container">
					{#if showing[i]}
						<svg
							xmlns="http://www.w3.org/2000/svg"
							fill="none"
							viewBox="0 0 24 24"
							stroke-width="1.5"
							stroke="currentColor"
							class="size-6"
						>
							<path stroke-linecap="round" stroke-linejoin="round" d="m4.5 15.75 7.5-7.5 7.5 7.5" />
						</svg>
					{:else}
						<svg
							xmlns="http://www.w3.org/2000/svg"
							fill="none"
							viewBox="0 0 24 24"
							stroke-width="1.5"
							stroke="currentColor"
							class="size-6"
						>
							<path stroke-linecap="round" stroke-linejoin="round" d="m19.5 8.25-7.5 7.5-7.5-7.5" />
						</svg>
					{/if}
				</div>
			</button>
			<div class="answer" class:hidden={!showing[i]}>
				{@html item.answer}
			</div>
		</div>
	{/each}
</div>

<style scoped>
	.info-page-container {
		width: 100%;
		box-sizing: border-box;
	}
	h1 {
		margin: 1.2em 0 !important;
		font-size: 2em;
		font-weight: 700;
	}
	.hidden {
		display: none;
	}

	.question-btn {
		border-radius: var(--radius-md);
		background: var(--light-background);
		border: 1px solid var(--border);
		color: var(--faded-text);
		padding: 1em 2rem;
		font-size: 1em;
		text-align: left;
		width: 100%;
		cursor: pointer;
		display: flex;
	}
	svg {
		width: 1.8em;
		opacity: 0.5;
	}
	.answer {
		padding: 2em 3rem;
		color: var(--subtle-text);
		font-size: 0.95em;
		overflow-wrap: break-word;
		overflow: hidden;
	}
	:global(.answer .shiki) {
		font-size: 0.85em;
		padding: 1.4em 2em;
		border-radius: 0.3em;
		overflow-x: auto;
	}
</style>
