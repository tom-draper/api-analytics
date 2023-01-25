<script lang="ts">
  import { onMount } from "svelte";

  // Integer to method string mapping used by server
  let methodMap = ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "CONNECT", "HEAD", "TRACE"];

  function endpointFreq(): {[endpointID: string]: {path: string, status: number, count: number}} {
    let freq = {};
    for (let i = 0; i < data.length; i++) {
      // Create groups of endpoints by path + status
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

    // Convert object to list
    let freqArr = [];
    maxCount = 0;
    for (let endpointID in freq) {
      freqArr.push(freq[endpointID]);
      if (freq[endpointID].count > maxCount) {
        maxCount = freq[endpointID].count;
      }
    }

    // Sort by count
    freqArr.sort((a, b) => {
      return b.count - a.count;
    });
    endpoints = freqArr;

    // Hide endpoint labels that don't fit inside bar once rendered
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
    if (endpoint.clientWidth < endpointCount.clientWidth) {
      endpointCount.style.display = "none";
    }
  }
  function setEndpointLabels() {
    for (let i = 0; i < endpoints.length; i++) {
      setEndpointLabelVisibility(i);
    }
  }

  let endpoints: any[];
  let maxCount: number;
  let mounted = false;
  onMount(() => {
    mounted = true;
  });

  $: data && mounted && build();

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
            title={endpoint.count}
            style="width: {(endpoint.count / maxCount) * 100}%"
            class:success={endpoint.status >= 200 && endpoint.status <= 299}
            class:bad={endpoint.status >= 300 && endpoint.status <= 399}
            class:error={endpoint.status >= 400 && endpoint.status <= 499}
          >
            <div class="endpoint-label" id="endpoint-label-{i}">
              <div class="path" id="endpoint-path-{i}">
                {endpoint.path}
              </div>
              <div class="count" id="endpoint-count-{i}">{endpoint.count.toLocaleString()}</div>
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

<style scoped>
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
    white-space: nowrap;
  }
  .endpoint-container {
    display: flex;
  }
  .external-label {
    padding: 3px 15px;
    left: 40px;
    top: 0;
    margin: 5px 0;
    color: var(--dim-text);
    display: none;
    font-size: 0.85em;
  }
  .success {
    background: var(--highlight);
  }
  .bad {
    background: rgb(235, 235, 129);
  }
  .error {
    background: var(--red);
  }
</style>
