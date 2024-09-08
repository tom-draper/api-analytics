<script lang="ts">
	export let reason: 'no-requests' | 'error';
</script>

<div class="container">
	<div
		class="hanging-lightning"
		class:hanging-lightning-no-requests={reason === 'no-requests'}
		class:hanging-lightning-error={reason === 'error'}
	>
		<div
			class="lightning-container"
			class:lightning-container-no-requests={reason === 'no-requests'}
			class:lightning-container-error={reason === 'error'}
		>
			{#if reason === 'no-requests'}
				<img src="../img/logo.png" alt="" />
			{:else}
				<img src="../img/logo-red.png" alt="" />
			{/if}
		</div>
	</div>
	<div class="message-container">
		{#if reason === 'no-requests'}
			<div class="message message-no-requests">No requests logged.</div>
		{:else}
			<div class="message message-error">Server error.</div>
		{/if}
	</div>
</div>

<style scoped>
	.container {
		color: var(--highlight);
		margin-bottom: 14em;
	}

	.hanging-lightning {
		width: 50%;
		height: 22em;
		position: relative;
		border-width: 1px;
		border-style: solid;
		border-top: none;
		border-left: none;
		border-bottom: none;
	}
	.hanging-lightning-no-requests {
		border-image: linear-gradient(#000, rgba(63, 207, 142, 0.35)) 30;
	}
	.hanging-lightning-error {
		border-image: linear-gradient(#000, rgba(228, 97, 97, 0.35)) 30;
	}

	.lightning-container {
		position: absolute;
		right: -22em;
		aspect-ratio: 1/1;
		height: 44em;
		display: grid;
		place-items: center;
	}
	.lightning-container-no-requests {
		background: radial-gradient(
			rgba(63, 207, 142, 0.3),
			transparent,
			transparent
		);
	}
	.lightning-container-error {
		background: radial-gradient(
			rgba(228, 97, 97, 0.3),
			transparent,
			transparent
		);
	}

	img {
		width: 20px;
	}

	.message-container {
		/* background: radial-gradient(circle closest-corner at center 125px, #222, transparent 20%) no-repeat; */
		/* background: radial-gradient(circle closest-corner at center 125px, black, transparent 20%) no-repeat; */
		text-align: center;
	}
	.message {
		font-weight: 700;
		display: inline-block;
		padding-top: 100px;
		padding-bottom: 20px;
		padding-right: 4em;
		padding-left: 4em;
		margin-top: -2em;
		width: fit-content;
		-webkit-background-clip: text;
		background-clip: text;
		-webkit-text-fill-color: transparent;
		margin-bottom: 6em;
		filter: saturate(1.3);
	}
	.message-no-requests {
		/* background: radial-gradient(circle farthest-corner at center center, var(--highlight), #333) no-repeat; */
		background: radial-gradient(
				circle farthest-corner at center center,
				var(--highlight),
				#111
			)
			no-repeat;
		background-clip: text;
	}
	.message-error {
		/* background: radial-gradient(circle farthest-corner at center center, var(--red), #333) no-repeat; */
		background: radial-gradient(
				circle farthest-corner at center center,
				var(--red),
				#111
			)
			no-repeat;
		background-clip: text;
	}
</style>
