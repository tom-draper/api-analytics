<script lang="ts">
    let generatedKey = false;
    let apiKey = "";
  async function genAPIKey() {
    if (!generatedKey) {
        const response = await fetch(
        "https://fastapi-analytics.vercel.app/api/gen-api-key"
        );
        const data = await response.json();
        console.log(data);
        if (data.status == 200) {
            generatedKey = true;
            apiKey = data.api-key
        }
    }
  }

  function copyToClipboard() {
    console.log("Copy");
  }
</script>

<div class="generate">
  <div class="content">
    <h2>Generate API key</h2>
    <input type="text" readonly bind:value={apiKey} >
    <button id="generateBtn" on:click={genAPIKey}>Generate</button>
    <button id="copyBtn" on:click={genAPIKey}>Generate</button>

    <div class="logo">API Analytics</div>
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
  .logo {
    color: #3fcf8e;
    margin-top: 3em;
    font-size: 0.8em;
  }
  #copyBtn {
    background: #1c1c1c;
    display: none;
  }
  #generateBtn {
    background: #3fcf8e;
  }
</style>
