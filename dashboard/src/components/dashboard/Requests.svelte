<script lang="ts">
  import { onMount } from "svelte";

  function setPercentageChange() {
    if (prevData.length == 0) {
      percentageChange = null;
    } else {
      percentageChange = (data.length / prevData.length) * 100 - 100;
    }
  }

  let percentageChange: number;
  let mounted = false;
  onMount(() => {
    mounted = true;
  });

  $: data && mounted && setPercentageChange();

  export let data: RequestsData, prevData: RequestsData;
</script>

<div class="card" title="Total">
  {#if percentageChange != null}
    <div
      class="percentage-change"
      class:positive={percentageChange > 0}
      class:negative={percentageChange < 0}
    >
      ({percentageChange > 0 ? "+" : ""}{percentageChange.toFixed(1)}%)
    </div>
  {/if}
  <div class="card-title">Requests</div>
  <div class="value">{data.length.toLocaleString()}</div>
</div>

<style>
  .card {
    width: calc(200px - 1em);
    margin: 0 1em 0 2em;
    position: relative;
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
</style>
