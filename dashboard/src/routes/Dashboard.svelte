<script lang="ts">
  import { onMount } from "svelte";

  async function fetchData() {
    // Fetch page ID
    const response = await fetch(
    //   `http://localhost:8080/data/${userID}`
      `https://api-analytics-server.vercel.app/api/data/${userID}`
    );
    console.log(response);
    if (response.status == 200) {
      data = await response.json();
      console.log(data)
    //   window.location.href = `/dashboard/${data.value}`;
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
  <div class="id">{userID}</div>
</div>
