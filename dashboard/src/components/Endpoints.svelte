<script lang="ts">
  import { onMount } from "svelte";

  function build() {
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
  function setEndpointLabelVisibility(idx: number) {
    let endpoint = document.getElementById(`endpoint-label-${idx}`)
    let endpointPath = document.getElementById(`endpoint-path-${idx}`)
    let endpointCount = document.getElementById(`endpoint-count-${idx}`)
    let externalLabel = document.getElementById(`external-label-${idx}`)
    if (endpoint.clientWidth < endpointPath.clientWidth + endpointCount.clientWidth) {
        externalLabel.style.display = 'flex';
        endpointPath.style.display = 'none';
    }
  }
  function setEndpointLabels() {
    for (let i = 0; i < endpoints.length; i++) {
        setEndpointLabelVisibility(i)
    }
  }
  onMount(() => {
    build();
    setTimeout(setEndpointLabels, 0);
  });
  let endpoints: any[];
  let maxCount: number;
  export let data: any;
</script>

<div class="card">
  <div class="card-title">Endpoints</div>
  {#if endpoints != undefined}
    <div class="endpoints">
      {#each endpoints as endpoint, i}
        <div class="endpoint-container">
          <div
            class="endpoint"
            id="endpoint-{i}"
            style="width: {(endpoint.count / maxCount) * 100}%"
          >
            <div class="endpoint-label" id="endpoint-label-{i}">
              <div class="path" id="endpoint-path-{i}">
                {endpoint.path}
              </div>
              <div class="count" id="endpoint-count-{i}">{endpoint.count}</div>
            </div>
          </div>
          <div class="external-label" id="external-label-{i}">
            <div class="external-label-path">
              {endpoint.path}
            </div>
          </div>
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
    position: relative;
  }
  .endpoint-label {
    display: flex;
  }
  .path,
  .count {
    padding: 3px 15px;
  }
  .path {
    flex-grow: 1;
  }
  .endpoint-container {
    display: flex;
  }
  .external-label {
    padding: 3px 15px;
    left: 40px;
    top: 0;
    margin: 10px 0;
    color: #707070;
    display: none;
  }
  .external-label-count {
    margin-left: 10px;
  }
</style>
