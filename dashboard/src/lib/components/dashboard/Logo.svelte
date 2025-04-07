<script lang="ts">
	import Lightning from "$components/Lightning.svelte";

	export let loading: boolean;
</script>

<div class="card">
	{#if loading}
		<div class="placeholder">
			<div class="spinner">
				<div class="loader" ></div>
			</div>
		</div>
	{:else}
		<div class="lightning">
			<Lightning />
		</div>
	{/if}
</div>

<style scoped>
	.card {
		width: calc(215px - 1em);
		margin: 0 1em 0 0;
		background: transparent;
		display: grid;
		place-items: center;
	}
	.lightning {
		width: 25px;
		color: var(--highlight);
	}
	.spinner {
		height: auto !important;
	}
	.loader {
		height: 36px !important;
		width: 36px !important;
	}
	@media screen and (max-width: 1030px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 1em 0 0;
		}
	}
</style>
