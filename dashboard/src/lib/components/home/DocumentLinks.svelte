<div class="docs m-auto flex w-fit pb-16 pt-4">
	<div class="doc-link-container flex-1 grid place-items-center">
		<a class="doc-link flex-1 bg-[#181818] hover:bg-[var(--hover-bg)]" href="/faq">FAQ</a>
	</div>
	<div class="doc-link-container flex-1 grid place-items-center">
		<a class="doc-link flex-1 bg-[#181818] hover:bg-[var(--hover-bg)]" href="/privacy-policy">Privacy Policy</a>
	</div>
	<div class="doc-link-container flex-1 grid place-items-center">
		<a class="doc-link flex-1 bg-[#181818] hover:bg-[var(--hover-bg)]" href="/outages">Outages</a>
	</div>
</div>

<style scoped>
	.doc-link {
		color: #5f5f5f;
		border-radius: var(--radius-md);
		min-width: 10em;
		max-width: 15em;
		margin: auto;
		padding: 0.8em 1.5em;
		margin: 0 1em;
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
