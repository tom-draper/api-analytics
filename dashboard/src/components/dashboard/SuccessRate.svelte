<script lang="ts">
  import { onMount } from "svelte";

  function build() {
    let totalRequests = 0;
    let successfulRequests = 0;
    for (let i = 0; i < data.length; i++) {
      if (data[i].status >= 200 && data[i].status <= 299) {
        successfulRequests++;
      }
      totalRequests++;
    }
    if (totalRequests > 0) {
      successRate = (successfulRequests / totalRequests) * 100;
    } else {
      successRate = 100;
    }
  }

  let mounted = false;
  let successRate: number;
  onMount(() => {
    mounted = true;
  });

  $: data && mounted && build();

  export let data: RequestsData;
</script>

<div class="card">
  <div class="card-title">Success rate</div>
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

<style scoped>
  .card {
    width: calc(200px - 1em);
    margin: 0 0 2em 1em;
  }
  .value {
    margin: 20px 0;
    font-size: 1.8em;
    font-weight: 600;
    color: var(--yellow);
  }
  @media screen and (max-width: 940px) {
    .card {
      width: auto;
      flex: 1;
    }
  }
</style>
