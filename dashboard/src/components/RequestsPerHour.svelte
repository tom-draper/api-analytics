<script lang="ts">
  import { onMount } from "svelte";
  function periodToDays(period: string): number {
    if (period == '24-hours') {
      return 1
    } else if (period == 'week') {
      return 8
    } else if (period == 'month') {
      return 30
    } else if (period == '3-months') {
      return 30*3
    } else if (period == '6-months') {
      return 30*6
    } else if (period == 'year') {
      return 365
    } else {
      return null
    }
  }
  function build() {
    let totalRequests = 0;
    for (let i = 0; i < data.length; i++) {
      totalRequests++;
    }
    console.log(totalRequests);
    if (totalRequests > 0) {
      let days = periodToDays(period);
      if (days != null) {
        requestsPerHour = (totalRequests / (24 * days)).toFixed(2);
      }
    } else {
      requestsPerHour = '0';
    }
  }

  let requestsPerHour: string;
  onMount(() => {
    build();
  });

  $: data && build();

  export let data: RequestsData, period: string;
</script>

<div class="card" title="Last week">
  <div class="card-title">
    Requests <span class="per-hour">/ hour</span>
  </div>
  {#if requestsPerHour != undefined}
    <div class="value">{requestsPerHour}</div>
  {/if}
</div>

<style>
  .card {
    width: calc(200px - 1em);
    margin: 0 2em 0 1em;
  }
  .value {
    margin: 20px 0;
    font-size: 1.8em;
    font-weight: 600;
  }
  .per-hour {
    color: var(--dim-text);
    font-size: 0.8em;
    margin-left: 4px;
  }
</style>
