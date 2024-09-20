<script lang="ts">
	import faq from '../lib/faq';
	import Footer from '../components/Footer.svelte';

	function toggleAnswer(index: number) {
		faq[index].showing = !faq[index].showing;
	}
</script>

<div class="info-page-container">
	<h1>Frequently Asked Questions</h1>

	{#each faq as question, i}
		<div class="faq-container">
			<button
				class="question-btn"
				on:click={() => {
					toggleAnswer(i);
				}}
			>
				<div class="question-container">
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
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								d="m4.5 15.75 7.5-7.5 7.5 7.5"
							/>
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
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								d="m19.5 8.25-7.5 7.5-7.5-7.5"
							/>
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
<Footer />

<style scoped>
	h1 {
		margin: 1.2em 0 !important;
	}
	.hidden {
		display: none;
	}

	.faq-container {
		margin-bottom: 1.5em;
	}

	.question-btn {
		border-radius: 4px;
		background: var(--light-background);
		border: 1px solid #2e2e2e;
		color: #ededed;
		padding: 1.2em 3rem;
		font-size: 1.2em;
		text-align: left;
		min-width: 100%;
		cursor: pointer;
		display: flex;
	}
	.question-container {
		flex: 1;
		margin: auto;
	}
	svg {
		width: 1.8em;
		opacity: 0.5;
	}
	.answer {
		margin-bottom: 1.5em;
	}
	.answer {
		padding: 2em 3rem;
		color: #707070;
		color: #c3c3c3;
	}

</style>
