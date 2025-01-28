<script lang="ts">
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
		<img src="/images/logos/lightning-green.png" alt="" />
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
	img {
		width: 25px;
	}
	.spinner {
		height: auto !important;
	}
	.loader {
		height: 20px !important;
		width: 20px !important;
	}
	@media screen and (max-width: 1030px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 1em 0 0;
		}
	}
</style>
