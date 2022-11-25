<script lang="ts">
  let generatedKey = false;
  let apiKey = "";
  let generateBtn: HTMLButtonElement;
  let copyBtn: HTMLButtonElement;
  async function genAPIKey() {
    if (!generatedKey) {
      const response = await fetch(
        // "http://localhost:8080/generate-api-key"
        "https://api-analytics-server.vercel.app/api/generate-api-key"
      );
      console.log(response.status);
      if (response.status == 200) {
        const data = await response.json();
        generatedKey = true;
        apiKey = data.value;
        generateBtn.style.display = "none";
        copyBtn.style.display = "initial";
      }
    }
  }

  function copyToClipboard() {
    navigator.clipboard.writeText(apiKey);
    copyBtn.value = "Copied";
  }
</script>

<div class="generate">
  <div class="content">
    <h2>Generate API key</h2>
    <input type="text" readonly bind:value={apiKey} />
    <button id="generateBtn" on:click={genAPIKey} bind:this={generateBtn}
      >Generate</button
    >
    <button id="copyBtn" on:click={copyToClipboard} bind:this={copyBtn}
      >Copy</button
    >

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
    /* color: #7e7e7e; */
    margin-bottom: 1em;
  }
  #copyBtn {
    background: #1c1c1c;
    display: none;
  }
  #generateBtn {
    background: #3fcf8e;
  }
</style>
