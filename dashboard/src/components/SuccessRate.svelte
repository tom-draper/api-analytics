<script lang="ts">
  import { onMount } from "svelte";

  function thisWeek(date: Date): boolean {
    let weekAgo = new Date()
    weekAgo.setDate(weekAgo.getDate() - 7);
    return date > weekAgo
  }

  function build() {
    let totalRequests = 0;
    let successfulRequests = 0;
    for (let i = 0; i < data.length; i++) {
      let date = new Date(data[i].created_at);
      if (thisWeek(date)) {
        if (data[i].status_code >= 200 && data[i].status_cde <= 299) {
          successfulRequests++
        }
        totalRequests++
      }
    }
    successRate = (successfulRequests/ totalRequests).toFixed(1)
  }

  let successRate: string;
  onMount(() => {
    build();
  });

  export let data: any;
</script>

<div class="card" title="Last month">
  <div class="card-title">
    Success Rate
  </div>
  {#if successRate != undefined}
    <div class="value">{successRate}%</div>
  {/if}
</div>

<style>
  .card {
    width: calc(200px - 1em);
    margin: 0 0 0 1em;
  }

  .value {
    margin: 20px 0;
    font-size: 2em;
    font-weight: 600;
  }

</style>
