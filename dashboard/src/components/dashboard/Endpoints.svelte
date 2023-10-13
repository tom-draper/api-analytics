<script lang="ts">
  import { onMount } from 'svelte';
  import { METHOD, PATH, STATUS } from '../../lib/consts';

  // Integer to method string mapping used by server
  const methodMap = [
    'GET',
    'POST',
    'PUT',
    'PATCH',
    'DELETE',
    'OPTIONS',
    'CONNECT',
    'HEAD',
    'TRACE',
  ];

  function endpointFreq(): {
    [endpointID: string]: { path: string; status: number; count: number };
  } {
    const freq = {};
    for (let i = 1; i < data.length; i++) {
      // Create groups of endpoints by path + status
      const endpointID = `${data[i][PATH]}${data[i][STATUS]}`;
      if (!(endpointID in freq)) {
        freq[endpointID] = {
          path: `${methodMap[data[i][METHOD]]}  ${data[i][PATH]}`,
          status: data[i][STATUS],
          count: 0,
        };
      }
      freq[endpointID].count++;
    }
    return freq;
  }

  function statusMatch(status: number): boolean {
    return (
      activeBtn === 'all' ||
      (activeBtn === 'success' && status >= 200 && status <= 299) ||
      (activeBtn === 'bad' && status >= 300 && status <= 399) ||
      (activeBtn === 'error' && status >= 400)
    );
  }

  function setTargetEndpoint(endpoint: string) {
    if (endpoint === targetEndpoint) {
      targetEndpoint = null;
    } else {
      targetEndpoint = endpoint;
    }
  }

  function build() {
    const freq = endpointFreq();

    // Convert object to list
    const freqArr = [];
    maxCount = 0;
    for (const endpointID in freq) {
      if (statusMatch(freq[endpointID].status)) {
        freqArr.push(freq[endpointID]);
        if (freq[endpointID].count > maxCount) {
          maxCount = freq[endpointID].count;
        }
      }
    }

    // Sort by count
    freqArr.sort((a, b) => {
      return b.count - a.count;
    });
    endpoints = freqArr.slice(0, 50);
  }

  function setEndpointLabelVisibility(idx: number) {
    const endpoint = document.getElementById(`endpoint-label-${idx}`);
    const endpointPath = document.getElementById(`endpoint-path-${idx}`);
    const endpointCount = document.getElementById(`endpoint-count-${idx}`);
    const externalLabel = document.getElementById(`external-label-${idx}`);
    if (
      endpoint.clientWidth <
      endpointPath.clientWidth + endpointCount.clientWidth
    ) {
      externalLabel.style.display = 'flex';
      endpointPath.style.display = 'none';
    }
    if (endpoint.clientWidth < endpointCount.clientWidth) {
      endpointCount.style.display = 'none';
    }
  }

  function setBtn(value: string) {
    activeBtn = value;
    build();
  }

  let endpoints: any[];
  let maxCount: number;
  let mounted = false;
  let activeBtn = 'all';
  onMount(() => {
    mounted = true;
  });

  $: data && mounted && build();

  export let data: RequestsData, targetEndpoint: string;
</script>

<div class="card">
  <div class="card-title">
    Endpoints
    <div class="toggle">
      <button
        class="cancel"
        class:visible={targetEndpoint != null}
        on:click={() => {
          setTargetEndpoint(null);
        }}>Cancel</button
      >
      <button
        class:active={activeBtn === 'all'}
        on:click={() => {
          setBtn('all');
        }}>All</button
      >
      <button
        class:active={activeBtn === 'success'}
        on:click={() => {
          setBtn('success');
        }}>Success</button
      >
      <button
        class:bad-active={activeBtn === 'bad'}
        on:click={() => {
          setBtn('bad');
        }}>Bad</button
      >
      <button
        class:error-active={activeBtn === 'error'}
        on:click={() => {
          setBtn('error');
        }}>Error</button
      >
    </div>
  </div>

  {#if endpoints != undefined}
    <div class="endpoints">
      {#each endpoints as endpoint, i}
        <div class="endpoint-container">
          <div
            class="endpoint"
            id="endpoint-{i}"
            title="Status: {endpoint.status}"
            on:click={() => setTargetEndpoint(endpoint.path.split(' ')[2])}
          >
            <div class="path">
              <b>{endpoint.count.toLocaleString()}</b>
              {endpoint.path}
            </div>
            <div
              class="background"
              style="width: {(endpoint.count / maxCount) * 100}%"
              class:success={(endpoint.status >= 200 &&
                endpoint.status <= 299) ||
                endpoint.status === 0}
              class:bad={endpoint.status >= 300 && endpoint.status <= 399}
              class:error={endpoint.status >= 400}
            />
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
  .card-title {
    display: flex;
  }
  .toggle {
    margin-left: auto;
  }
  .active {
    background: var(--highlight);
  }
  .bad-active {
    background: rgb(235, 235, 129);
  }
  .error-active {
    background: var(--red);
  }
  button {
    border: none;
    border-radius: 4px;
    background: rgb(68, 68, 68);
    cursor: pointer;
    padding: 2px 6px;
    margin-left: 5px;
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
    width: 100%;
    cursor: pointer;
  }
  .path {
    position: relative;
    flex-grow: 1;
    z-index: 1;
    pointer-events: none;
    color: #505050;
    padding: 3px 12px;
  }
  .endpoint-container {
    display: flex;
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
  .cancel {
    display: none;
    background: gold;
  }
  .visible {
    display: inline;
  }
  .background {
    border-radius: 3px;
    color: var(--light-background);
    text-align: left;
    position: relative;
    font-size: 0.85em;
    height: 100%;
    position: absolute;
    top: 0;
  }
  @media screen and (max-width: 1030px) {
    .card {
      width: auto;
      flex: 1;
      margin: 0 0 2em 0;
    }
  }
</style>
