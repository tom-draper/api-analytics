<script lang="ts">
  import { onMount } from 'svelte';

  function selectOption(option: string) {
    selected = option;
    hideOptions = true;
  }

  function setWidth() {
    dropdown.style.width = `${current.clientWidth}px`;
  }

  let dropdown: HTMLDivElement;
  let current: HTMLButtonElement;
  onMount(() => {
    setWidth();
    setWidth();  // Sometimes needs to be called twice when first loaded in mobile view (maybe due to image render?)
  });

  let hideOptions: boolean = true;
  export let options: string[], selected: string;
</script>

<div class="dropdown" bind:this={dropdown} id="dropdown">
  <div class="inner">
    <button
      class="current"
      class:square-bottom={!hideOptions}
      bind:this={current}
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
      {selected}
    </button>
    <div class="options" class:hidden={hideOptions}>
      {#each options as option, i}
        {#if option !== selected}
          <button
            class="option"
            class:last-option={i === options.length - 1}
            on:click={() => {
              selectOption(option);
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
    padding: 5px 12px 5px 9px;
    cursor: pointer;
    display: flex;
  }
  .options {
    display: flex;
    flex-direction: column;
    border-radius: 0px 4px 4px 0px;
    background: var(--background);
    color: var(--dim-text);
    top: 66px;
    z-index: 100;
  }
  .option {
    background: var(--background);
    color: var(--dim-text);
    border: 1px solid #2e2e2e;
    padding: 5px 12px;
    cursor: pointer;
  }
  .hidden {
    display: none;
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
    width: inherit;
    flex-direction: column;
  }
  svg {
    width: 16px;
    margin-right: 5px;
    opacity: 0.6;
  }
</style>
