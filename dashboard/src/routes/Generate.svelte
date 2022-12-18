<script lang="ts">
  let loading = false;
  let generatedKey = false;
  let apiKey = "";
  let generateBtn: HTMLButtonElement;
  let copyBtn: HTMLButtonElement;
  let copiedNotification: HTMLDivElement;
  async function genAPIKey() {
    if (!generatedKey) {
      loading = true;
      const response = await fetch(
        "https://api-analytics-server.vercel.app/api/generate-api-key"
      );
      if (response.status == 200) {
        const data = await response.json();
        generatedKey = true;
        apiKey = data.value;
        generateBtn.style.display = "none";
        copyBtn.style.display = "grid";
      }
      loading = false;
    }
  }

  function copyToClipboard() {
    navigator.clipboard.writeText(apiKey);
    copyBtn.value = "Copied";
    copiedNotification.style.visibility = "visible";
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
      ><img class="copy-icon" src="img/copy.png" alt="" /></button
    >
    <div id="copied" bind:this={copiedNotification}>Copied!</div>

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

<style>
  .generate {
    display: grid;
    place-items: center;
  }
  h2 {

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
    font-size: 0.8em;
  }
  .keep-secure {
    color: #5a5a5a;
    margin-bottom: 1em;
  }

  #copyBtn {
    background: #1c1c1c;
    display: none;
    background: #343434;
    place-items: center;
    margin: auto;
  }
  .copy-icon {
    filter: contrast(0.3);
    height: 20px;
  }
  #copied {
    color: var(--highlight);
    margin: 2em auto auto;
    visibility: hidden;
    height: 1em;
  }

  .spinner {
    height: 7em;
    /* margin-bottom: 5em; */
  }
  #generateBtn {
    background: #3fcf8e;
  }
</style>
