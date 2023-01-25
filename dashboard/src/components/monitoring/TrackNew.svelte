<script lang="ts">
  async function postMonitor() {
    try {
      let response = await fetch(
        `https://api-analytics-server.vercel.app/api/monitor/add`,
        {
          method: "POST",
          mode: "no-cors",
          headers: {
            Accept: "application/json",
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            api_key: apiKey,
            url: url,
            ping: pingType == "simple-ping",
            secure: false 
          }),
        }
      );
      if (response.status != 201) {
        console.log("Error", response.status);
      }
      showTrackNew = false;
    } catch (e) {
      console.log(e);
    }
  }

  function setPingType(value: string) {
    pingType = value;
  }

  let url: string;
  let pingType = "simple-ping";

  export let apiKey: string, showTrackNew: boolean;
</script>

<div class="card">
  <div class="card-text">
    <div class="url">
      <div class="start">http://</div>
      <input type="text" placeholder="example.com/endpoint/" bind:value="{url}" />
      <button class="add" on:click={postMonitor}>Add</button>
    </div>
    <div class="ping-types">
      <button
        class="ping-type"
        on:click={() => setPingType("simple-ping")}
        class:active={pingType == "simple-ping"}
      >
        Ping
      </button>
      <button
        class="ping-type"
        on:click={() => setPingType("get-request")}
        class:active={pingType == "get-request"}
      >
        GET
      </button>
    </div>
    <div class="detail">
      Endpoints are pinged by our servers every 30 mins and response <b
        >status</b
      >
      and response <b>time</b> are logged.
    </div>
  </div>
</div>

<style scoped>
  .card {
    width: min(100%, 1000px);
    border: 1px solid #2e2e2e;
    margin: 2.2em auto;
  }
  .card-text {
    margin: 2em 2em 1.9em;
  }
  input {
    background: var(--background);
    border-radius: 4px;
    border: none;
    margin: 0 10px;
    width: 100%;
    text-align: left;
    height: auto;
    padding: 6px 15px;
    color: white;
  }
  .url {
    display: flex;
  }
  .start {
    margin: auto;
    color: var(--dim-text);
  }
  .detail {
    margin-top: 30px;
    color: var(--dim-text);
    font-weight: 400;
    font-size: 0.85em;
  }
  button {
    border: none;
    border-radius: 4px;
    background: var(--light-background);
    cursor: pointer;
  }
  .add {
    background: var(--highlight);
    padding: 5px 20px;
  }
  .ping-types {
    margin-top: 20px;
  }
  .ping-type {
    color: var(--dim-text);
    border: 1px solid #2e2e2e;
    padding: 5px 20px;
    width: 80px;
  }
  .active {
    background: var(--highlight);
    color: black;
  }
</style>
