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
      buildStats();
    }
  }



  function buildStats() {

  }

  let data: any;
  let stats: any;
  onMount(() => {
    fetchData();
  });
  export let userID: string;
</script>

<div>
  <h1>Dashboard</h1>
  <div class="id">{userID}</div>
  {#if data != undefined}
    <ResponseTimes {data} />
    {/if}
</div>
