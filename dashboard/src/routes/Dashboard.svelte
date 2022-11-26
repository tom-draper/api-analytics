<script lang="ts">
  import { onMount } from "svelte";
  import Requests from "../components/Requests.svelte";
  import ResponseTimes from "../components/ResponseTimes.svelte";
  import Endpoints from "../components/Endpoints.svelte";
  import Footer from "../components/Footer.svelte";
  import SuccessRate from "../components/SuccessRate.svelte";
  import PastMonth from "../components/PastMonth.svelte";
  import Browser from "../components/Browser.svelte";
  import OperatingSystem from "../components/OperatingSystem.svelte";

  function formatUUID(userID: string): string {
    return `${userID.slice(0, 8)}-${userID.slice(8, 12)}-${userID.slice(
      12,
      16
    )}-${userID.slice(16, 20)}-${userID.slice(20)}`;
  }

  async function fetchData() {
    userID = formatUUID(userID);
    // Fetch page ID
    const response = await fetch(
      `https://api-analytics-server.vercel.app/api/data/${userID}`
    );
    if (response.status == 200) {
      const json = await response.json();
      data = json.value;
      console.log(data);
    }
  }

  let data: any;
  onMount(() => {
    fetchData();
  });
  export let userID: string;
</script>

<div>
  <!-- <h1>Dashboard</h1> -->
  {#if data != undefined}
    <div class="dashboard">
      <div class="left">
        <div class="row">
          <Requests {data} />
          <SuccessRate {data} />
        </div>
        <ResponseTimes {data} />
        <Endpoints {data} />
      </div>
      <div class="right">
        <PastMonth {data} />
        <div class="row">

          <OperatingSystem {data} />
          <Browser {data} />
        </div>
      </div>
    </div>
    {/if}
  <Footer />
</div>

<style>
  .dashboard {
    margin: 5em;
    display: flex;
  }
  .row {
    display: flex;
  }
  .right {
    flex-grow: 1;
  }
</style>