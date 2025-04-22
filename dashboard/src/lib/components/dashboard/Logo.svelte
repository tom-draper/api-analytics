<script lang="ts">
	import Lightning from '$components/Lightning.svelte';
	import Loading from '$components/Loading.svelte';

	export let loading: boolean;
</script>

<div class="card">
	<div class="lightning">
		{#if loading}
			<Loading />
		{:else}
			<Lightning />
		{/if}
	</div>
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
		height: 44px;
		color: var(--highlight);
	}
	@media screen and (max-width: 1030px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 1em 0 0;
		}
	}
</style>
