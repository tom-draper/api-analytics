<script lang="ts">
  import { onMount } from "svelte";

  function pastWeek(date: Date): boolean {
    let weekAgo = new Date();
    weekAgo.setDate(weekAgo.getDate() - 7);
    return date > weekAgo;
  }

  function build() {
    let totalRequests = 0;
    let successfulRequests = 0;
    for (let i = 0; i < data.length; i++) {
      let date = new Date(data[i].created_at);
      if (pastWeek(date)) {
        if (data[i].status >= 200 && data[i].status <= 299) {
          successfulRequests++;
        }
        totalRequests++;
      }
    }
    successRate = (successfulRequests / totalRequests) * 100;
  }

  let successRate: number;
  onMount(() => {
    build();
  });

  export let data: any;
</script>

<div class="card" title="Last week">
  <div class="card-title">Success Rate</div>
  {#if successRate != undefined}
    <div
      class="value"
      style="color: {successRate <= 75 ? 'var(--red)' : ''}{successRate > 75 &&
      successRate < 90
        ? 'var(--yellow)'
        : ''}{successRate >= 90 ? 'var(--highlight)' : ''}"
    >
      {successRate.toFixed(1)}%
    </div>
  {/if}
</div>

<style>
  .card {
    width: calc(200px - 1em);
    margin: 0 0 0 1em;
  }

  .value {
    margin: 20px 0;
    font-size: 1.8em;
    font-weight: 600;
    color: var(--yellow);
  }
</style>
