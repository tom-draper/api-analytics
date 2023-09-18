<script lang="ts">
    function selectOption(option: string) {
        selected = option
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
    }
    .current {
        border-radius: 4px;
        background: var(--background);
    }
    .hidden {
        display: none;
    }
    .active {
        background: var(--highlight);
        flex: 1;
    }
</style>