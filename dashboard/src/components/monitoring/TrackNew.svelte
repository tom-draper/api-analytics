<script lang="ts">
  import type {NotificationState} from '../../lib/notification';
  import Dropdown from '../dashboard/Dropdown.svelte';

  function triggerErrorMessage(message: string) {
    notification.message = message
    notification.style = "error"
    notification.show = true
    setTimeout(() => {
      notification.show = false;
    }, 4000)
  }

  async function postMonitor() {
    if (url == null) {
      triggerErrorMessage("URL is blank.")
      return
    } else if (monitorCount >= 3) {
      triggerErrorMessage("Maximum 3 monitors allowed.")
      return
    }

    try {
      const response = await fetch(
        `https://www.apianalytics-server.com/api/monitor/add`,
        {
          method: 'POST',
          mode: 'no-cors',
          headers: {
            Accept: 'application/json',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            user_id: userID,
            url: url,
            ping: true,
            secure: false,
          }),
        }
      );
      if (response.status !== 201) {
        console.log('Error', response.status);
      }
      showTrackNew = false;
    } catch (e) {
      console.log(e);
    }
  }

  let url: string;
  let options = ['https', 'http'];
  let urlPrefix = options[0];

  export let userID: string,
    showTrackNew: boolean,
    monitorCount: number,
    notification: NotificationState;
</script>

<div class="card">
  <div class="card-text">
    <div class="url">
      <Dropdown {options} selected={urlPrefix} />
      <input
        type="text"
        placeholder="www.example.com/endpoint/"
        bind:value={url}
      />
      <button class="add" on:click={postMonitor}>Add</button>
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
    margin: 1px 10px;
    width: 100%;
    text-align: left;
    height: auto;
    padding: 4px 15px;
    color: white;
    font-size: 0.9em;
  }
  .url {
    display: flex;
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
    padding: 4px 20px;
    margin: 1px 0;
  }
</style>
