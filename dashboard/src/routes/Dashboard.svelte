<script lang="ts">
  import { onMount } from "svelte";
  import Requests from "../components/Requests.svelte";
  import Welcome from "../components/Welcome.svelte";
  import RequestsPerHour from "../components/RequestsPerHour.svelte";
  import ResponseTimes from "../components/ResponseTimes.svelte";
  import Endpoints from "../components/Endpoints.svelte";
  import Footer from "../components/Footer.svelte";
  import SuccessRate from "../components/SuccessRate.svelte";
  import PastMonth from "../components/PastMonth.svelte";
  import Version from "../components/Version.svelte";
  import UsageTime from "../components/UsageTime.svelte";
  import Growth from "../components/Growth.svelte";
  import Device from "../components/Device.svelte";

  function formatUUID(userID: string): string {
    return `${userID.slice(0, 8)}-${userID.slice(8, 12)}-${userID.slice(
      12,
      16
    )}-${userID.slice(16, 20)}-${userID.slice(20)}`;
  }

  async function fetchData() {
    userID = formatUUID(userID);
    // Fetch page ID
    try {
      const response = await fetch(
        `https://api-analytics-server.vercel.app/api/data/${userID}`
      );
      if (response.status == 200) {
        const json = await response.json();
        data = json.value;
        console.log(data);
      }
    } catch (e) {
      failed = true;
    }
  }

  let data: RequestsData;
  let failed = false;
  onMount(() => {
    fetchData();
  });
  export let userID: string;
</script>

{#if data != undefined}
  <div class="dashboard">
    <div class="left">
      <div class="row">
        <Welcome />
        <SuccessRate {data} />
      </div>
      <div class="row">
        <Requests {data} />
        <RequestsPerHour {data} />
      </div>
      <ResponseTimes {data} />
      <Endpoints {data} />
      <Version {data} />
    </div>
    <div class="right">
      <PastMonth {data} />
      <div class="grid-row">
        <Growth {data} />
        <Device {data} />
      </div>
      <UsageTime {data} />
    </div>
  </div>
{:else if failed}
  <div class="no-requests">No requests currently logged.</div>
{:else}
  <div class="placeholder" style="min-height: 85vh;">
    <div class="spinner">
      <div class="loader"/>
    </div>

  </div>
{/if}
<Footer />

<style>
  .dashboard {
    min-height: 90vh;
  }
  .dashboard {
    margin: 5em;
    display: flex;
  }
  .row {
    display: flex;
  }
  .grid-row {
    display: flex;
  }
  .right {
    flex-grow: 1;
    margin-right: 2em;
  }
  .no-requests {
    height: 70vh;
    font-size: 1.5em;
    display: grid;
    place-items: center;
    color: var(--highlight);
  }
  .placeholder {
    min-height: 85vh;
    display: grid;
    place-items: center;
  }
  @media screen and (max-width: 1580px){
    .grid-row {
        flex-direction: column;
      }
  }
  @media screen and (max-width: 1100px) {
    .dashboard {
      margin: 2em 0;
    }
  }
</style>
