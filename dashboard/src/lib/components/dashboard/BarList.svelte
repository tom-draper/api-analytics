<script lang="ts">
	export type BarRow = {
		value: string;
		label: string;
		count: number;
		height: number;
		selected: boolean;
	};

	let {
		title,
		rows,
		onSelect,
		minHeight = '300px',
		marginTop = '0'
	}: {
		title: string;
		rows: BarRow[];
		onSelect: (value: string) => void;
		minHeight?: string;
		marginTop?: string;
	} = $props();
</script>

<div class="card" style="--bar-min-height: {minHeight}; --bar-margin-top: {marginTop};">
	<div class="card-title">{title}</div>
	<div class="list">
		{#each rows as row (row.value)}
			<div class="row-container">
				<button
					class="row"
					class:selected={row.selected}
					onclick={() => onSelect(row.value)}
				>
					<div class="label">
						<span class="font-semibold">{row.count.toLocaleString()}</span>
						{row.label}
					</div>
					<div class="background" style="width: {row.height * 100}%"></div>
				</button>
			</div>
		{/each}
	</div>
</div>

<style scoped>
	.card {
		margin-left: 2em;
		margin-top: var(--bar-margin-top);
		min-height: var(--bar-min-height);
	}
	.list {
		margin: 0.9em 20px 0.6em;
	}
	.row {
		border-radius: var(--radius-sm);
		margin: 5px 0;
		color: var(--light-background);
		text-align: left;
		position: relative;
		font-size: 0.85em;
		width: 100%;
		cursor: pointer;
	}
	.row:hover {
		background: var(--fade-right);
	}
	.selected {
		background: var(--fade-right);
	}
	.label {
		position: relative;
		flex-grow: 1;
		z-index: 1;
		pointer-events: none;
		color: var(--muted-text);
		padding: 3px 12px;
		overflow-wrap: break-word;
	}
	.row-container {
		display: flex;
	}
	.background {
		border-radius: var(--radius-sm);
		background: var(--highlight);
		text-align: left;
		position: absolute;
		top: 0;
		height: 100%;
		font-size: 0.85em;
	}
	@media screen and (max-width: 1600px) {
		.card {
			margin-left: 0;
			width: 100%;
			min-height: unset;
		}
	}
	@media screen and (max-width: 1070px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 0 2em 0;
		}
	}
</style>
