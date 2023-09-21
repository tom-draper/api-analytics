<script lang="ts">
  import { onMount } from "svelte";

    function selectOption(option: string) {
        selected = option
        hideOptions = true;
    }

    let dropdown: HTMLDivElement;
    let current: HTMLButtonElement;
    onMount(() => {
        dropdown.style.width = `${current.clientWidth}px`;
    })

    let hideOptions: boolean = true;
    export let options: string[], selected: string;
</script>

<div class="dropdown" bind:this={dropdown} id="dropdown">
    <div class="inner">
        <button class="current" class:square-bottom={!hideOptions} bind:this={current} on:click={() => {hideOptions = !hideOptions}}>
            {selected}
        </button>
        <div class="options" class:hidden={hideOptions}>
            {#each options as option, i}
                {#if option !== selected}
                    <button class="option" class:last-option={i === options.length - 1} on:click={() => {selectOption(option)}}>{option}</button>
                {/if}
            {/each}
        </div>
    </div>
</div>

<style scoped>
    .dropdown {
        display: flex;
        flex-direction: column;
        margin-right: 10px;
    }
    .current {
        border-radius: 4px;
        background: var(--background);
        color: var(--dim-text);
        border: 1px solid #2e2e2e;
        padding: 5px 12px;
    }
    .options {
        display: flex;
        flex-direction: column;
        /* position: absolute; */
        border-radius: 0px 4px 4px 0px;
        background: var(--background);
        color: var(--dim-text);
        /* border: 1px solid #2e2e2e; */
        /* padding: 5px 12px; */
        /* margin-right: 10px; */
        top: 66px;
        z-index: 100;
    }
    .option {
        /* border-radius: 4px; */
        background: var(--background);
        color: var(--dim-text);
        border: 1px solid #2e2e2e;
        padding: 5px 12px;
        /* margin-right: 10px; */
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
        position: absolute;
        z-index: 9;
        display: flex;
        flex-direction: column;
    }
    /* .dropdown {
        width: 120px;
    } */
</style>