<script lang="ts">
	function selectOption(option: string | null) {
		selected = option;
		hideOptions = true;
	}

	let hideOptions: boolean = true;
	export let options: string[],
		selected: string | null,
		defaultOption: string;
</script>

<div class="dropdown" id="dropdown">
	<div class="inner">
		<button
			class="current"
			class:square-bottom={!hideOptions}
			on:click={() => {
				hideOptions = !hideOptions;
			}}
		>
			<svg
				xmlns="http://www.w3.org/2000/svg"
				fill="none"
				viewBox="0 0 24 24"
				stroke-width="2"
				stroke="currentColor"
				class="w-6 h-6"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					d="M19.5 8.25l-7.5 7.5-7.5-7.5"
				/>
			</svg>
			{selected || defaultOption}
		</button>
		<div class="options" class:hidden={hideOptions}>
			{#each [defaultOption, ...options] as option, i}
				{#if option !== selected && option !== null && (selected !== null || option !== defaultOption)}
					<button
						class="option"
						class:last-option={(selected === defaultOption &&
							i === options.length - 1) ||
							(selected !== defaultOption &&
								i === options.length)}
						on:click={() => {
							if (option === defaultOption) {
								selectOption(null);
							} else {
								selectOption(option);
							}
						}}>{option}</button
					>
				{/if}
			{/each}
		</div>
	</div>
</div>

<style scoped>
	.dropdown {
		display: flex;
		flex-direction: column;
		height: 28px;
	}
	.current {
		border-radius: 4px;
		background: var(--background);
		color: var(--dim-text);
		border: 1px solid #2e2e2e;
		padding: 5px 15px 5px 9px;
		cursor: pointer;
		display: flex;
		pointer-events: all;
	}
	.options {
		display: flex;
		flex-direction: column;
		border-radius: 0px 0 4px 0px;
		background: var(--background);
		color: var(--dim-text);
		top: 66px;
		z-index: 100;
		margin-bottom: 50px;
		width: 100%;
	}
	.option {
		background: var(--background);
		color: var(--dim-text);
		border: 1px solid #2e2e2e;
		border-top: none;
		padding: 5px 15px 5px 31px;
		text-align: left;
		cursor: pointer;
	}
	.hidden {
		visibility: hidden;
	}
	.last-option {
		border-radius: 0 0 4px 4px;
	}
	.square-bottom {
		border-radius: 4px 4px 0 0;
	}
	.inner {
		z-index: 9;
		display: flex;
		width: inherit;
		flex-direction: column;
		pointer-events: none;
	}
	svg {
		width: 16px;
		margin-right: 6px;
		opacity: 0.6;
	}
</style>
