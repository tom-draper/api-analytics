<script lang="ts">
    function selectOption(option: string) {
        selected = option
        hideOptions = true;
    }

    let hideOptions: boolean = true;
    export let options: string[], selected: string;
</script>

<div class="dropdown">
    <button class="current" on:click={() => {hideOptions = !hideOptions}}>
        {selected}
    </button>
    <div class="options" class:hidden={hideOptions}>
        {#each options as option}
            {#if option !== selected}
                <button class="option" on:click={() => {selectOption(option)}}>{option}</button>
            {/if}
        {/each}
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
        position: absolute;
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
        display: none;
    }

</style>