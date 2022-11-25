<script lang="ts">
  import { onMount } from "svelte";
  import ResponseTimes from "../components/ResponseTimes.svelte";

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
    }
  }

  let data: any;
  onMount(() => {
    fetchData();
  });
  export let userID: string;
</script>

<div>
  <h1>Dashboard</h1>
  {#if data != undefined}
    <div class="dashboard">
      <ResponseTimes {data} />
    </div>
    {/if}
</div>

<style>
  .dashboard {
    margin: 5em;
  }
</style>