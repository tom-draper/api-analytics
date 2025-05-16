<script lang="ts">
	export let generic: boolean = false;
</script>

<div class="footer grid place-items-center mt-[1.5em] mb-[3em]">
	{#if generic}
		<a
			class="github-link"
			rel="noreferrer"
			target="_blank"
			href="https://github.com/tom-draper/api-analytics"
		>
			<img class="github-logo mb-[30px] h-[28px]" src="/images/logos/github.png" alt="" />
		</a>
	{:else}
		<div class="mb-[2em] text-[0.8em] text-[#5a5a5a]">Keep your API key safe and secure.</div>
	{/if}
	<div class="logo text-[13px] text-[var(--highlight)]">API Analytics</div>
	<img class="footer-logo" src="/images/logos/lightning-green.svg" alt="" />
</div>

<style scoped>
	.github-logo {
		filter: invert(0.24);
		transition: filter 0.15s;
	}
	.github-logo:hover {
		filter: invert(0.5);
	}
</style>
