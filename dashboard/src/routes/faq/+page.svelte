<script lang="ts">
	import faq from '$lib/faq';

	function toggleAnswer(index: number) {
		faq[index].showing = !faq[index].showing;
	}
</script>

<svelte:head>
	<link rel="stylesheet" href="/css/prism.css" />
</svelte:head>

<div class="info-page-container">
	<h1>Frequently Asked Questions</h1>

	{#each faq as question, i}
		<div class="mb-4">
			<button
				class="question-btn"
				on:click={() => {
					toggleAnswer(i);
				}}
			>
				<div class="m-auto flex-1">
					{question.question}
				</div>
				<div class="dropdown-icon-container">
					{#if question.showing}
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
			<div class="answer" class:hidden={!question.showing}>
				{@html question.answer}
			</div>
		</div>
	{/each}
</div>

<style scoped>
	h1 {
		margin: 1.2em 0 !important;
		font-size: 2em;
		font-weight: 700;
	}
	.hidden {
		display: none;
	}

	.question-btn {
		border-radius: 4px;
		background: var(--light-background);
		border: 1px solid #2e2e2e;
		color: #ededed;
		padding: 1em 2rem;
		font-size: 1em;
		text-align: left;
		min-width: 100%;
		cursor: pointer;
		display: flex;
	}
	svg {
		width: 1.8em;
		opacity: 0.5;
	}
	.answer {
		padding: 2em 3rem;
		color: #c3c3c3;
		font-size: 0.95em;
	}
</style>
