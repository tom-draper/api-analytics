<script lang="ts">
  let apiKey = "";
  let loading = false
  let message = '';
  async function genAPIKey() {
    loading = true;
    // Fetch page ID
    const response = await fetch(
      `https://api-analytics-server.vercel.app/api/delete/${apiKey}`
    );

    if (response.status == 200) {
      message = 'Deleted successfully'
    } else {
      message = 'Error: API key invalid'
    }
    loading = false;
  }
</script>

<div class="generate">
  <div class="content">
    <h2>Delete all stored data</h2>
    <input type="text" bind:value={apiKey} placeholder="Enter API key"/>
    <button id="generateBtn" on:click={genAPIKey}>Delete</button>
    <div class="notification">{message}</div>
    <div class="spinner">
      <div class="loader" style="display: {loading ? 'initial' : 'none'}" />
    </div>
  </div>
  <div class="details">
    <!-- <div class="keep-secure">Keep your API key safe and secure.</div> -->
    <div class="highlight logo">API Analytics</div>
    <img class="footer-logo" src="img/logo.png" alt="">
  </div>
</div>

<style>
  .generate {
    display: grid;
    place-items: center;
  }
  h2 {
    /* font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, */
      /* Ubuntu, Cantarell, "Open Sans", "Helvetica Neue", sans-serif; */
    margin: 0 0 1em;
    font-size: 2em;
  }
  .content {
    width: fit-content;
    background: #343434;
    background: #1c1c1c;
    padding: 3.5em 4em 4em;
    border-radius: 9px;
    margin: 20vh 0 2vh;
    height: 400px;
  }
  input {
    background: #1c1c1c;
    background: #343434;
    border: none;
    padding: 0 20px;
    width: 310px;
    font-size: 1em;
    text-align: center;
    height: 40px;
    border-radius: 4px;
    margin-bottom: 2.5em;
    color: white;
    display: grid;
  }
  button {
    height: 40px;
    border-radius: 4px;
    padding: 0 20px;
    border: none;
    cursor: pointer;
    width: 100px;
  }
  .highlight {
    color: #3fcf8e;
  }
  .details {
    /* margin-top: 15rem; */
    font-size: 0.8em;
    margin-top: calc(15px + 1em);
  }
  .notification {
    color: #3fcf8e;
    margin-top: 32px;
    height: 16px;
  }
  #generateBtn {
    background: #3fcf8e;
  }
</style>
