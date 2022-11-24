<script lang="ts">
  async function genAPIKey() {
    // Fetch page ID
    const response = await fetch(
      "https://api-analytics-server.vercel.app/api/data",
      {
        method: 'POST',
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(userID)
      }
    );
    
    console.log(response)
    if (response.status == 200) {
      const data = await response.json();
      window.location.href = `/dashboard/${data.value}`;
    }
  }
    export let userID: string;
</script>

<div>
    <h1>Dashboard</h1>
    <div class="id">{userID}</div>
</div>
