<script lang="ts">
  import { onMount } from "svelte";
  import Footer from "../components/Footer.svelte";
  import Card from "../components/monitoring/Card.svelte";
  import TrackNew from "../components/monitoring/TrackNew.svelte";

  function formatUUID(userID: string): string {
    return `${userID.slice(0, 8)}-${userID.slice(8, 12)}-${userID.slice(
      12,
      16
    )}-${userID.slice(16, 20)}-${userID.slice(20)}`;
  }

  async function fetchData() {
    userID = formatUUID(userID);
    try {
      const response = await fetch(
        `https://api-analytics-server.vercel.app/api/pings/${userID}`
      );
      if (response.status == 200) {
        const json = await response.json();
        data = json.monitor;
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

  type MonitorData = {
	  api_key: string,
    url: string,
    secure: boolean,
    ping: boolean,
    created_at: Date
  }

  let error = false;
  let period = "30d";
  let data: {monitor: MonitorData[], api_key: string};
  let measurements = Array(3);
  let failed = false;

  for (let i = 0; i < measurements.length; i++) {
    measurements[i] = {
      name: "persona-api.vercel.app/v1/england",
      measurements: [],
    };
    for (let j = 0; j < 140; j++) {
      measurements[i].measurements.push({
        status: "success",
        response_time: Math.random() * 10 + 5,
      });
    }
  }

  for (let i = 50; i < 58; i++) {
    measurements[0].measurements[i] = { status: "error", response_time: 0 };
  }
  measurements[1].name = "persona-api.vercel.app/v1/england/features";

  let showTrackNew = false;
  onMount(() => {
    fetchData();
  });

  export let userID: string;
</script>

<div class="monitoring">
  <div class="status">
    {#if error}
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
    {#if showTrackNew || measurements.length == 0}
      <TrackNew api_key={data.api_key}/>
    {/if}
    <Card data={measurements[0]} {period} bind:anyError={error} />
    <Card data={measurements[1]} {period} bind:anyError={error} />
    <Card data={measurements[2]} {period} bind:anyError={error} />
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
    width: 130px;
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
</style>
