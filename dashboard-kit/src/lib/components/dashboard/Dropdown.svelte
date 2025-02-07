<script lang="ts">
	import { onMount } from 'svelte';

	let dropdown: HTMLDivElement;

	function selectOption(option: string | null) {
		selected = option;
		open = true;
	}

	function closeDropdown(e) {
		// Check if the click is outside the dropdown
		if (!dropdown.contains(e.target)) {
			open = false;
		}
	}

	function toggleOpen() {
		open = !open;
	}

	onMount(() => {
		document.addEventListener('click', closeDropdown);

		return () => {
			document.removeEventListener('click', closeDropdown);
		};
	});

	export let open: boolean = false;
	export let options: string[], selected: string | null, defaultOption: string;
</script>

<div class="dropdown" id="dropdown" bind:this={dropdown}>
	<div class="inner" class:no-click={!open}>
		<button class="current" class:square-bottom={open} on:click={toggleOpen}>
			<svg
				xmlns="http://www.w3.org/2000/svg"
				fill="none"
				viewBox="0 0 24 24"
				stroke-width="2"
				stroke="currentColor"
				class="h-6 w-6"
			>
				<path stroke-linecap="round" stroke-linejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
			</svg>
			{selected || defaultOption}
		</button>
		<div class="options" class:hidden={!open}>
			{#each [defaultOption, ...options] as option, i}
				{#if option !== selected && option !== null && (selected !== null || option !== defaultOption)}
					<button
						class="option"
						class:last-option={(selected === defaultOption && i === options.length - 1) ||
							(selected !== defaultOption && i === options.length)}
						on:click={() => {
							const value = option === defaultOption ? null : option;
							selectOption(value);
							open = false;
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
		padding: 4px 15px 4px 9px;
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
	.option:hover,
	.current:hover {
		background: #161616;
	}
	.current:hover svg {
		transform: translateY(1px); /* Moves icon 3px down */
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
	}
	.no-click {
		pointer-events: none;
	}
	svg {
		width: 16px;
		height: 16px;
		align-self: center;
		margin-right: 6px;
		opacity: 0.6;
		display: inline-block;

		transition: transform 0.15s ease-in-out;
	}
</style>
