<script lang="ts">
	function removeItem(item: string) {
		const newItems: Set<string> = new Set();
		for (const value of items) {
			if (value !== item) {
				newItems.add(value);
			}
		}
		items = newItems;
	}

	function addItem(item: string) {
		const newItems = new Set(items);
		if (item.charAt(0) !== '/') {
			item = '/' + item;
		}
		newItems.add(item);
		items = newItems;
		input.value = null;
	}

	function handleInputKeyDown(e: KeyboardEvent) {
		if (e.keyCode === 13) {
			addItem(input.value);
		}
	}

	let input: HTMLInputElement;
	const hideOptions: boolean = true;

	export let items: Set<string>, placeholder: string;
</script>

<div class="container">
	<div class="inner">
		<input
			type="text"
			name="item"
			id=""
			{placeholder}
			bind:this={input}
			on:keydown={handleInputKeyDown}
		/>
		<div class="items" class:hidden={hideOptions}>
			{#each Array.from(items) as item, _}
				<div class="item">
					<button
						class="remove-btn"
						on:click={() => {
							removeItem(item);
						}}
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							fill="none"
							viewBox="0 0 24 24"
							stroke-width="1.5"
							stroke="currentColor"
							class="size-6"
						>
							<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
						</svg>
					</button>
					<div class="item-text">
						{item}
					</div>
				</div>
			{/each}
		</div>
	</div>
</div>

<style scoped>
	.container {
		display: flex;
		flex-direction: column;
		width: 100%;
	}
	.items {
		display: flex;
		flex-direction: column;
		border-radius: 0px 4px 4px 0px;
		background: var(--background);
		width: 100%;
	}
	.item {
		background: var(--background);
		color: var(--dim-text);
		border: 1px solid #2e2e2e;
		display: flex;
		border-radius: 3px;
	}
	.item-text {
		flex-grow: 1;
		text-align: left;
		align-content: center;
		color: #ededed !important;
		margin: 4px 12px;
		font-size: 0.85em;
	}
	input {
		margin: 10px 0;
		height: 35px;
		text-align: left;
		padding: 0.1em 1em;
		font-size: 0.9em;
		width: auto;
		background: #2b2b2b;
	}
	input::placeholder {
		color: #707070;
	}
	.remove-btn {
		background: transparent;
		outline: none;
		border: none;
		color: white;
		padding: 0 0.5em;
		cursor: pointer;
		border-radius: 2px;
		height: 35px;
		width: 35px;
	}
	.remove-btn:hover {
		background: rgb(35, 35, 35);
	}
	svg {
		width: 18px;
	}

	.inner {
		z-index: 9;
		display: flex;
		flex-direction: column;
		font-size: 1em;
	}
</style>
