<script lang="ts">
  import { onMount } from 'svelte';
  import Footer from '../components/Footer.svelte';
  import Card from '../components/monitoring/Card.svelte';
  import TrackNew from '../components/monitoring/TrackNew.svelte';
  import Notification from '../components/dashboard/Notification.svelte';
  import formatUUID from '../lib/uuid';
  import type { NotificationState } from '../lib/notification';
  import { serverURL } from '../lib/consts';

  async function fetchData() {
    userID = formatUUID(userID);
    try {
      const response = await fetch(`${serverURL}/api/monitor/pings/${userID}`);
      if (response.status === 200) {
        data = await response.json();
      }
    } catch (e) {
      console.log(e);
    }
  }

  function setPeriod(value: string) {
    period = value;
    error = false;
  }

  function toggleShowTrackNew() {
    showTrackNew = !showTrackNew;
  }

  function addEmptyMonitor(url: string) {
    data[url] = [];
  }

  function removeMonitor(url: string) {
    delete data[url];
    data = data; // Trigger reactivity to update display
  }

  let error = false;
  const periods = ['24h', '7d', '30d', '60d'];
  let period = periods[1];
  let data: MonitorData;
  let notification: NotificationState = {
    message: '',
    style: 'error',
    show: false,
  };

  let showTrackNew = false;
  onMount(async () => {
    await fetchData();
    console.log(data);
  });

  export let userID: string;
</script>

<div class="monitoring">
  <div class="status">
    {#if data !== undefined && Object.keys(data).length === 0}
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
        <button
          class="add-new-btn"
          class:active={showTrackNew}
          on:click={toggleShowTrackNew}
          ><div class="add-new-text">
            <span class="plus">+</span>
          </div>
        </button>
      </div>
      <div class="period-controls-container">
        <div class="period-controls">
          {#each periods as _period}
            <button
              class="period-btn"
              class:active={period === _period}
              on:click={() => {
                setPeriod(_period);
              }}
            >
              {_period}
            </button>
          {/each}
        </div>
      </div>
    </div>
    {#if data}
      {#if showTrackNew || Object.keys(data).length == 0}
        <TrackNew
          {userID}
          bind:showTrackNew
          monitorCount={Object.keys(data).length}
          bind:notification
          {addEmptyMonitor}
        />
      {/if}
      {#each Object.keys(data).sort() as url}
        <Card
          bind:url
          {data}
          {userID}
          {period}
          bind:anyError={error}
          bind:notification
          {removeMonitor}
        />
      {/each}
    {:else}
      <div class="spinner">
        <div class="loader" />
      </div>
    {/if}
  </div>
</div>
<Notification bind:state={notification} />
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
  .period-controls-container {
    margin-top: auto;
  }

  button {
    background: var(--light-background);
    color: var(--dim-text);
    border: none;
    padding: 3px 12px;
    cursor: pointer;
  }
  .add-new-btn:hover {
    background: radial-gradient(var(--light-background), #3fcf8e10);
  }
  .add-new-btn {
    border: 1px solid #2e2e2e;
    border-radius: 4px;
    padding: 0;
    height: 35px;
    color: var(--highlight);
    width: 35px;
  }
  .add-new-text {
    display: flex;
    font-size: 2em;
    justify-content: center;
  }
  .active,
  .active:hover {
    background: var(--highlight);
    color: black !important;
  }
  .spinner {
    margin: 3em 0 10em;
  }

  @media screen and (max-width: 1100px) {
    .cards-container {
      width: 95%;
    }
  }
</style>
