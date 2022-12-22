<script lang="ts">
  let apiKey = "";
  let loading = false;
  let message = "";
  async function genAPIKey() {
    loading = true;
    // Fetch page ID
    const response = await fetch(
      `https://api-analytics-server.vercel.app/api/delete/${apiKey}`
    );

    if (response.status == 200) {
      message = "Deleted successfully";
    } else {
      message = "Error: API key invalid";
    }
    loading = false;
  }
</script>

<div class="generate">
  <div class="content">
    <h2>Delete all stored data</h2>
    <input type="text" bind:value={apiKey} placeholder="Enter API key" />
    <button id="formBtn" on:click={genAPIKey}>Delete</button>
    <div class="notification">{message}</div>
    <div class="spinner">
      <div class="loader" style="display: {loading ? 'initial' : 'none'}" />
    </div>
  </div>
  <div class="details">
    <div class="keep-secure">Keep your API key safe and secure.</div>
    <div class="highlight logo">API Analytics</div>
    <img class="footer-logo" src="img/logo.png" alt="" />
  </div>
</div>

<style scoped>
  .notification {
    color: #3fcf8e;
    margin-top: 32px;
    height: 16px;
  }
</style>
