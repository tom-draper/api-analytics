<script lang="ts">
  import { onMount } from "svelte";
  import periodToDays from "../../lib/period";

  function setPercentageChange() {
    if (prevData.length == 0) {
      percentageChange = null;
    } else {
      percentageChange = (data.length / prevData.length) * 100 - 100;
    }
  }

  function setRequestsPerHour() {
    if (data.length > 0) {
      let days = periodToDays(period);
      if (days != null) {
        requestsPerHour = (data.length / (24 * days)).toFixed(2);
      }
    } else {
      requestsPerHour = "0";
    }
  }
  
  function togglePeriod() {
    perHour = !perHour;
  }
  
  function build() {
    setPercentageChange();
    setRequestsPerHour();
  }
  
  let requestsPerHour: string;
  let perHour = false;
  let percentageChange: number;
  let mounted = false;
  onMount(() => {
    mounted = true;
  });
  
  $: data && mounted && build();

  export let data: RequestsData, prevData: RequestsData, period: string;
</script>

{#if perHour}
  <button class="card" on:click="{togglePeriod}">
    <div class="card-title">
      Requests <span class="per-hour">/ hour</span>
    </div>
    {#if requestsPerHour != undefined}
      <div class="value">{requestsPerHour}</div>
    {/if}
  </button>
{:else}
  <button class="card" on:click="{togglePeriod}">
    {#if percentageChange != null}
      <div
      class="percentage-change"
      class:positive={percentageChange > 0}
      class:negative={percentageChange < 0}
      >
        {#if percentageChange > 0}
          <img class="arrow" src="../img/up.png" alt="" />
        {:else if percentageChange < 0}
          <img class="arrow" src="../img/down.png" alt="" />
        {/if}
        {percentageChange.toFixed(1)}%
      </div>
    {/if}
    <div class="card-title">Requests</div>
    <div class="value">{data.length.toLocaleString()}</div>
  </button>
{/if}

<style scoped>
  .card {
    width: calc(200px - 1em);
    margin: 0 1em 0 2em;
    position: relative;
    cursor: pointer;
  }
  .value {
    margin: 20px 0;
    font-size: 1.8em;
    font-weight: 600;
  }
  .percentage-change {
    position: absolute;
    right: 20px;
    top: 20px;
    font-size: 0.8em;
  }
  .positive {
    color: var(--highlight);
  }
  .negative {
    color: rgb(228, 97, 97);
  }

  .per-hour {
    color: var(--dim-text);
    font-size: 0.8em;
    margin-left: 4px;
  }
  button {
    font-size: unset;
    font-family: unset;
  }
  .arrow {
    height: 11px;
  }
</style>
