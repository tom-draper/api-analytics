<div class="docs flex w-fit m-auto pb-16">
	<div class="doc-link-container">
		<a class="doc-link" href="/frequently-asked-questions">FAQ</a>
	</div>
	<div class="doc-link-container">
		<a class="doc-link" href="/privacy-policy">Privacy Policy</a>
	</div>
	<div class="doc-link-container">
		<a class="doc-link" href="/outages">Outages</a>
	</div>
</div>

<style scoped>
	.doc-link-container {
		flex: 1;
		display: grid;
		place-items: center;
	}
	.doc-link {
		flex: 1;
		color: #5f5f5f;
		background: #181818;
		border-radius: 4px;
		min-width: 10em;
		max-width: 15em;
		margin: auto;
		padding: 0.8em 1.5em;
		display: flex;
		margin: 0 1em;
		text-align: center;
		justify-content: center;
	}

	a:hover {
		background: #161616;
	}

	@media (max-width: 800px) {
		.docs {
			width: 100%;
			flex-direction: column;
			font-size: 0.8em;
			margin-top: 1em;
		}
		.doc-link {
			margin-top: 1em;
			width: 100%;
		}
	}
</style>
