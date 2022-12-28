<script lang="ts">
  let apiKey = "";
  let loading = false
  async function genAPIKey() {
    loading = true;
    const response = await fetch(
      `https://api-analytics-server.vercel.app/api/user-id/${apiKey}`
    );

    if (response.status == 200) {
      const userID = await response.json();
      window.location.href = `/dashboard/${userID.replaceAll("-", "")}`;
    }
    loading = false;
  }
</script>

<div class="generate">
  <div class="content">
    <h2>Dashboard</h2>
    <input type="text" bind:value={apiKey} placeholder="Enter API key"/>
    <button id="formBtn" on:click={genAPIKey}>Load</button>
    <div class="spinner">
      <div class="loader" style="display: {loading ? 'initial' : 'none'}" />
    </div>
  </div>
  <div class="details">
    <div class="keep-secure">Keep your API key safe and secure.</div>
    <div class="highlight logo">API Analytics</div>
    <img class="footer-logo" src="img/logo.png" alt="">
  </div>
</div>
