<script lang="ts">
  import { onMount } from "svelte";
  import Footer from "../components/Footer.svelte";
  import Card from "../components/monitoring/Card.svelte";
  import TrackNew from "../components/monitoring/TrackNew.svelte";
  import formatUUID from "../lib/uuid";

  async function fetchData() {
    userID = formatUUID(userID);
    try {
      const response = await fetch(
        `https://api-analytics-server.vercel.app/api/pings/${userID}`
      );
      if (response.status == 200) {
        const json = await response.json();
        data = json.pings;
        apiKey = json.api_key;
        console.log(data);
      }
    } catch (e) {
      failed = true;
    }
  }

  function setPeriod(value: string) {
    period = value;
    error = false;
  }

  function toggleShowTrackNew() {
    showTrackNew = !showTrackNew;
  }

  function groupByUrl() {
    let group = {};
    for (let i = 0; i < data.length; i++) {
      if (!(data[i].url in group)) {
        group[data[i].url] = [];
      }
      group[data[i].url].push({
        status: data[i].status,
        responseTime: data[i].response_time,
      });
    }
    return group;
  }

  type PingsData = {
    url: string;
    response_time: number;
    status: number;
    created_at: Date;
  };

  type MonitorSample = {
    status: number;
    responseTime: number;
  };

  let error = false;
  let period = "30d";
  let apiKey: string;
  let data: PingsData[];
  let monitorData: { [url: string]: MonitorSample[] };
  let failed = false;

  let showTrackNew = false;
  onMount(async () => {
    await fetchData();
    monitorData = groupByUrl();
  });

  export let userID: string;
</script>

<div class="monitoring">
  <div class="status">
    {#if monitorData != undefined && Object.keys(monitorData).length == 0}
    <div class="status-image">
      <img id="status-image" src="/img/logo.png" alt="" />
      <div class="status-text">Setup required</div>
    </div>
    {:else if error}
    <div class="status-image">
      <img id="status-image" src="/img/bigcross.png" alt="" />
      <div class="status-text">Systems down</div>
    </div>
    {:else}
      <div class="status-image">
        <img id="status-image" src="/img/bigtick.png" alt="" />
        <div class="status-text">Systems Online</div>
      </div>
    {/if}
  </div>
  <div class="cards-container">
    <div class="controls">
      <div class="add-new">
        <button class="add-new-btn" on:click={toggleShowTrackNew}
          ><div class="add-new-text">
            <span class="plus">+</span> New
          </div>
        </button>
      </div>
      <div class="period-controls-container">
        <div class="period-controls">
          <button
            class="period-btn {period == '24h' ? 'active' : ''}"
            on:click={() => {
              setPeriod("24h");
            }}
          >
            24h
          </button>
          <button
            class="period-btn {period == '7d' ? 'active' : ''}"
            on:click={() => {
              setPeriod("7d");
            }}
          >
            7d
          </button>
          <button
            class="period-btn {period == '30d' ? 'active' : ''}"
            on:click={() => {
              setPeriod("30d");
            }}
          >
            30d
          </button>
          <button
            class="period-btn {period == '60d' ? 'active' : ''}"
            on:click={() => {
              setPeriod("60d");
            }}
          >
            60d
          </button>
        </div>
      </div>
    </div>
    {#if monitorData != undefined}
      {#if showTrackNew || Object.keys(monitorData).length == 0}
        <TrackNew {apiKey} />
      {/if}
      {#each Object.entries(monitorData) as [url, samples]}
        <Card {url} data={samples} {period} bind:anyError={error} />
      {/each}
    {:else}
      <div class="spinner">
        <div class="loader" />
      </div>
    {/if}
  </div>
</div>
<Footer />

<style scoped>
  .monitoring {
    font-weight: 600;
  }
  .status {
    margin: 13vh 0 9vh;
    display: grid;
    place-items: center;
  }
  #status-image {
    height: 130px;
    margin-bottom: 1em;
    filter: saturate(1.3);
  }
  .status-text {
    font-size: 2.2em;
    font-weight: 700;
    color: white;
  }

  .cards-container {
    width: 60%;
    margin: auto;
    padding-bottom: 1em;
  }

  .controls {
    margin: auto;
    width: 60%;

    width: min(100%, 1000px);
    display: flex;
  }
  .add-new {
    flex-grow: 1;
    display: flex;
    justify-content: left;
  }
  .period-controls {
    margin-left: auto;
    display: flex;
    justify-content: right;
  }

  .period-controls {
    border: 1px solid #2e2e2e;
    width: fit-content;
    border-radius: 4px;
    overflow: hidden;
  }

  button {
    background: #232323;
    color: #707070;
    border: none;
    padding: 3px 12px;
    cursor: pointer;
  }
  .add-new-btn {
    border: 1px solid #2e2e2e;
    border-radius: 4px;
  }
  .add-new-text {
    display: flex;
  }
  .active {
    background: var(--highlight);
    color: black;
  }
  .plus {
    padding-right: 0.6em;
  }
  .spinner {
    margin: 3em 0 10em;
  }
</style>
