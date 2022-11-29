<script lang="ts">
  import { onMount } from "svelte";

  function thisWeek(date: Date): boolean {
    let weekAgo = new Date();
    weekAgo.setDate(weekAgo.getDate() - 7);
    return date > weekAgo;
  }

  function build() {
    let totalRequests = 0;
    for (let i = 0; i < data.length; i++) {
      let date = new Date(data[i].created_at);
      if (thisWeek(date)) {
        totalRequests++;
      }
    }
    requestsPerHour = ((24 * 7) / totalRequests).toFixed(2);
  }

  let requestsPerHour: string;
  onMount(() => {
    build();
  });

  export let data: RequestsData;
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
    margin: 0 1em 0 2em;
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
