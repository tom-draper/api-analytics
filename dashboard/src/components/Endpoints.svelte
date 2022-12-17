<script lang="ts">
  import { onMount } from "svelte";

  let methodMap = ['GET', 'POST']

  function endpointFreq(): any {
    let freq = {};
    for (let i = 0; i < data.length; i++) {
      let endpointID = data[i].path + data[i].status;
      if (!(endpointID in freq)) {
        freq[endpointID] = {
          path: `${methodMap[data[i].method]}  ${data[i].path}`,
          status: data[i].status,
          count: 0,
        };
      }
      freq[endpointID].count++;
    }
    return freq;
  }


  function build() {
    let freq = endpointFreq();

    let freqArr = [];
    maxCount = 0;
    for (let endpointID in freq) {
      freqArr.push(freq[endpointID]);
      if (freq[endpointID].count > maxCount) {
        maxCount = freq[endpointID].count;
      }
    }

    freqArr.sort((a, b) => {
      return b.count - a.count;
    });
    endpoints = freqArr;
    
    setTimeout(setEndpointLabels, 50);
  }
  
  function setEndpointLabelVisibility(idx: number) {
    let endpoint = document.getElementById(`endpoint-label-${idx}`);
    let endpointPath = document.getElementById(`endpoint-path-${idx}`);
    let endpointCount = document.getElementById(`endpoint-count-${idx}`);
    let externalLabel = document.getElementById(`external-label-${idx}`);
    if (
      endpoint.clientWidth <
      endpointPath.clientWidth + endpointCount.clientWidth
    ) {
      externalLabel.style.display = "flex";
      endpointPath.style.display = "none";
    }
  }
  function setEndpointLabels() {
    for (let i = 0; i < endpoints.length; i++) {
      setEndpointLabelVisibility(i);
    }
  }
  onMount(() => {
    build();
  });
  let endpoints: any[];
  let maxCount: number;

  $: data && build();

  export let data: RequestsData;
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
            style="width: {(endpoint.count / maxCount) *
              100}%; background: {endpoint.status >= 200 &&
            endpoint.status <= 299
              ? 'var(--highlight)'
              : '#e46161'}"
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
  .card {
    min-height: 361px;
  }
  .endpoints {
    margin: 0.9em 20px 0.6em;
  }
  .endpoint {
    border-radius: 3px;
    margin: 5px 0;
    color: var(--light-background);
    text-align: left;
    position: relative;
    font-size: 0.85em;
  }
  .endpoint-label {
    display: flex;
  }
  .path,
  .count {
    padding: 3px 15px;
  }
  .count {
    margin-left: auto;
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
    margin: 5px 0;
    color: #707070;
    display: none;
    font-size: 0.85em;
  }
</style>
