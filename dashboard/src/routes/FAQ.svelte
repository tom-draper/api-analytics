<script lang="ts">
	import Footer from '../components/Footer.svelte';

	const questions = [
		{
			question: 'Can the project be self-hosted?',
			answer: 'We are currently working on a self-hosted solution that will allow you to take full control over your logged request data. Check back soon for updates!',
			showing: false,
		},
		{
			question: 'How is this project funded?',
			answer: 'This project is fully funded by myself as a hobby project.',
			showing: false,
		},
		{
			question: 'I have lost my API key, how can I recover it?',
			answer: 'Your API key is the only link between you and your logged requests. We do not store any other forms of authentication such as email addresses. If you have lost your API key, please get in contact with us, we may be able to recover it for you. If you no longer access your account, your account and all associated data will be deleted after 3 months, or 6 months if your API continues to log requests.',
			showing: false,
		},
		{
			question: 'How do you comply with GDPR?',
			answer: 'We adhere to GDPR regulations and requirements. Our privacy policy can be found here.',
			showing: false,
		},
		{
			question: 'What are the usage limits?',
			answer: 'In order to keep the service free, we can currently only store up to 1.5 million requests per user. When you hit this limit, new requests will replace old requests. If you find this is not enough, we recommend using another service or waiting until our self-hosting solution is ready.',
			showing: false,
		},
		{
			question: 'Why are there requests missing from my dashboard?',
			answer: 'In order to keep the service free, we can currently only store up to 1.5 million requests per user. If you have hit this limit, old requests will be replaced with new requests. Occassionally we will have an outage that is reflected in your dashboard. You can track outages here.',
			showing: false,
		},
		{
			question: 'Can I contribute to the project?',
			answer: 'The project is open source and contributions, feedback and suggestions are always welcome.',
			showing: false,
		},
		{
			question: 'How long do you store data for?',
			answer: 'Data is only stored for as long as the service may be in use. Any account along with all associated data is scheduled for deletion when either: 3 months have elapsed since the last request was logged by your API, or 6 months have elapsed since you accessed the dashboard or data API, whichever happens first.',
			showing: false,
		},
		{
			question: 'How can I delete my logged request data?',
			answer: 'You can delete your account and all associated data at any time by entering your API key here.',
		},
		{
			question: 'How can I amend or modify my logged request data?',
			answer: 'We do not support the ability for users to amend or modify their logged request data themselves. If you have a specific request, please get in contact with us.',
		},
		{
			question: 'How can I request my data?',
			answer: 'You can access your raw logged request data by sending a GET request to our data API.',
		}
	];

	function toggleAnswer(index: number) {
		questions[index].showing = !questions[index].showing;
	}
</script>

<div class="frequently-asked-questions-container">
	<h1>Frequently Asked Questions</h1>

	{#each questions as question, i}
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
				{question.answer}
			</div>
		</div>
	{/each}
</div>
<Footer />

<style scoped>
	.frequently-asked-questions-container {
		max-width: 800px;
		margin: auto;
		padding: 4em 2em;
		text-align: left;
		color: white;
		font-family: 'Noto Sans' sans-serif;
	}
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
		width: 2em;
		opacity: 0.5;
	}
	.answer {
		margin-bottom: 1.5em;
	}
	.answer {
		padding: 2em 3rem;
		color: #707070;
	}
</style>
