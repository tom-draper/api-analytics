<script lang="ts">
	export let status: number, message: string;
</script>

<div class="min-h-[75vh] overflow-hidden text-[var(--highlight)]">
	<div
		class="hanging-lightning"
		class:hanging-lightning-no-requests={status !== 500}
		class:hanging-lightning-error={status === 500}
	>
		<div
			class="lightning-container"
			class:lightning-container-no-requests={status !== 500}
			class:lightning-container-error={status === 500}
		>
			{#if status === 500}
				<img src="/images/logos/lightning-red.png" alt="" />
			{:else}
				<img src="/images/logos/lightning-green.png" alt="" />
			{/if}
		</div>
	</div>
	<div class="message-container relative">
		{#if status !== 400 && status !== 500}
			<div
				class="absolute top-0 mt-[-0.6em] w-full text-center text-3xl font-bold !text-[var(--highlight)]"
			>
				<div class="mr-1 flex justify-center">
					<span class="pr-4">4</span>
					<span class="pl-4">4</span>
				</div>
			</div>
		{/if}
		<div
			class="message"
			class:message-no-requests={status !== 500}
			class:message-error={status === 500}
		>
			{#if status === 400}
				No requests found.
			{:else if status === 500}
				Internal server error.
			{:else}
				Page not found.
			{/if}
		</div>
	</div>
	<div class="description">{message}</div>
</div>

<style scoped>
	.description {
		color: var(--dim-text);
		padding: 7em 0 1em;
		font-size: 0.8em;
		min-height: 10em;
		display: grid;
		place-items: center;
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
		border-image: linear-gradient(#000, rgba(63, 207, 142, 0.7)) 30;
	}
	.hanging-lightning-error {
		border-image: linear-gradient(#000, rgba(228, 97, 97, 0.7)) 30;
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
		background: radial-gradient(rgba(63, 207, 142, 0.3), transparent, transparent);
	}
	.lightning-container-error {
		background: radial-gradient(rgba(228, 97, 97, 0.3), transparent, transparent);
	}

	img {
		width: 20px;
	}

	.message-container {
		text-align: center;
	}
	.message {
		font-weight: 700;
		display: inline-block;
		padding-top: 100px;
		padding-bottom: 100px;
		padding-right: 2em;
		padding-left: 2em;
		margin-top: -2em;
		width: fit-content;
		-webkit-background-clip: text;
		background-clip: text;
		-webkit-text-fill-color: transparent;
		filter: saturate(1.3);
	}
	.message-no-requests {
		background: radial-gradient(circle farthest-corner at center center, var(--highlight), #111)
			no-repeat;
		background-clip: text;
	}
	.message-error {
		background: radial-gradient(circle farthest-corner at center center, var(--red), #111) no-repeat;
		background-clip: text;
	}
</style>
