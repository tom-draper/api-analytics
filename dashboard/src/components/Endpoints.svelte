<script lang="ts">
  import { onMount } from "svelte";

  function setEndpoints() {
    let eps = {};
    for (let i = 0; i < data.length; i++) {
      if (!(data[i].path in eps)) {
        eps[data[i].path] = { count: 0 };
      }
      eps[data[i].path].count++;
    }

    let _endpoints = [];
    maxCount = 0;
    for (let path in eps) {
      _endpoints.push({ path: path, count: eps[path].count });
      if (eps[path].count > maxCount) {
        maxCount = eps[path].count;
      }
    }
    _endpoints.sort((a, b) => {
      return b.count - a.count;
    });
    endpoints = _endpoints;
  }
  onMount(() => {
    setEndpoints();
  });
  let endpoints: any[];
  let maxCount: number;
  export let data: any;
</script>

<div class="card">
  <div class="card-title">Endpoints</div>
  {#if endpoints != undefined}
    <div class="endpoints">
      {#each endpoints as endpoint, _}
        <div
          class="endpoint"
          style="width: {(endpoint.count / maxCount) * 100}%"
        >
          <div class="path">
            {endpoint.path}
          </div>
          <div class="count">{endpoint.count}</div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .endpoints {
    margin: 1em 20px 20px;
  }
  .endpoint {
    background: var(--highlight);
    border-radius: 4px;
    margin: 10px 0;
    color: var(--light-background);
    text-align: left;
    display: flex;
  }
  .path,
  .count {
    padding: 3px 15px;
  }
  .path {
    flex-grow: 1;
  }
</style>
