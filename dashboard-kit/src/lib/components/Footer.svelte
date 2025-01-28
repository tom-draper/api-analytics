<script lang="ts">
	export let generic: boolean = false;
</script>

<div class="footer">
	{#if generic}
		<a
			class="github-link"
			rel="noreferrer"
			target="_blank"
			href="https://github.com/tom-draper/api-analytics"
		>
			<img class="github-logo" height="30px" src="/images/logos/github.png" alt="" />
		</a>
	{:else}
		<div class="text-[#5a5a5a] mb-[2em] text-[0.8em]">Keep your API key safe and secure.</div>
	{/if}
	<div class="logo">API Analytics</div>
	<img class="footer-logo" src="/images/logos/lightning-green.svg" alt="" />
</div>

<style scoped>
	.logo {
		font-size: 13px;
		color: var(--highlight);
	}
	.footer {
		margin: 1.5em 0 3em;
		display: grid;
		place-items: center;
	}
	.github-logo {
		height: 28px;
		filter: invert(0.24);
		margin-bottom: 30px;
		transition: filter 0.15s;
	}
	.github-logo:hover {
		filter: invert(0.5);
	}
</style>
