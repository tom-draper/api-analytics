<script lang="ts">
  let apiKey = "";
  async function genAPIKey() {
    // Fetch page ID
    const response = await fetch(
      "https://fastapi-analytics.vercel.app/api/user-id",
      {
        method: 'POST',
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(apiKey)
      }
    );
    
    if (response.status == 200) {
      const data = await response.json();
      window.location.href = `/dashboard/${data.value}`;
    }
  }
</script>

<div class="generate">
  <div class="content">
    <h2>Enter API key:</h2>
    <input type="text" bind:value={apiKey} />
    <button id="generateBtn" on:click={genAPIKey}>Load</button>
    <div class="details">
      <div class="keep-secure">Keep your API key safe and secure.</div>
      <div class="highlight logo">API Analytics</div>
    </div>
  </div>
</div>

<style>
  .generate {
    display: grid;
    place-items: center;
  }
  h2 {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen,
      Ubuntu, Cantarell, "Open Sans", "Helvetica Neue", sans-serif;
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
  }
  input {
    background: #1c1c1c;
    background: #343434;
    border: none;
    padding: 0 20px;
    width: 300px;
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
    margin-top: 15em;
    font-size: 0.8em;
  }
  .keep-secure {
    color: #5a5a5a;
    margin-bottom: 1em;
  }
  #generateBtn {
    background: #3fcf8e;
  }
</style>
