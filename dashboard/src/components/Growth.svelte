<script lang="ts">
  import { onMount } from "svelte";

  function pastWeek(date: Date): boolean {
    let weekAgo = new Date();
    weekAgo.setDate(weekAgo.getDate() - 7);
    return date > weekAgo;
  }

  function lastWeek(date: Date): boolean {
    let weekAgo = new Date();
    weekAgo.setDate(weekAgo.getDate() - 7);
    let fortnightAgo = new Date();
    fortnightAgo.setDate(weekAgo.getDate() - 14);
    return date < weekAgo && date > fortnightAgo;
  }

  function buildWeek() {
    let thisWeek = { requests: 0, success: 0, responseTime: 0 };
    let prevWeek = { requests: 0, success: 0, responseTime: 0 };
    for (let i = 0; i < data.length; i++) {
      let date = new Date(data[i].created_at);
      if (pastWeek(date)) {
        // @ts-ignore
        thisWeek.requests++;
        if (data[i].status >= 200 && data[i].status <= 299) {
          thisWeek.success++;
        }
        thisWeek.responseTime += data[i].response_time
      } else if (lastWeek(date)) {
        prevWeek.requests++;
        if (data[i].status >= 200 && data[i].status <= 299) {
          prevWeek.success++;
        }
        prevWeek.responseTime  += data[i].response_time
      }
    }

    requestsChange = (((thisWeek.requests + 1) / (prevWeek.requests + 1)) * 100);
    successChange = ((((thisWeek.success + 1) / (thisWeek.requests + 1)) + 1) / (((prevWeek.success + 1) / (prevWeek.requests + 1)) + 1)) * 100
    responseTimeChange = ((((thisWeek.responseTime + 1) / (thisWeek.requests + 1)) + 1) / (((prevWeek.responseTime + 1) / (prevWeek.requests + 1)) + 1)) * 100
  }

  let requestsChange: number;
  let successChange: number;
  let responseTimeChange: number;
  onMount(() => {
    buildWeek();
  });

  export let data: RequestsData;
</script>

<div class="card">
  <div class="card-title">Growth</div>
  <div class="values">
    {#if requestsChange != undefined}
        <div class="tile">
          <div class="tile-value">
            <span
              style="color: {requestsChange > 0
                ? 'var(--highlight)'
                : 'var(--red)'}">{requestsChange >= 0 ? '+' : '-'}{requestsChange.toFixed(1)}%</span
            >
          </div>
          <div class="tile-label">Requests</div>
        </div>
    {/if}
    <!-- {#if requestsChange != undefined}
        <div class="tile">
          <div class="tile-value">
            <span
              style="color: {requestsChange > 0
                ? 'var(--highlight)'
                : 'var(--red)'}">+{requestsChange}%</span
            >
          </div>
          <div class="tile-label">Users</div>
        </div>
    {/if} -->
    {#if successChange != undefined}
        <div class="tile">
          <div class="tile-value">
            <span
              style="color: {successChange > 0
                ? 'var(--highlight)'
                : 'var(--red)'}">{successChange >= 0 ? '+' : '-'}{successChange.toFixed(1)}%</span
            >
          </div>
          <div class="tile-label">Success rate</div>
        </div>
    {/if}
    {#if responseTimeChange != undefined}
        <div class="tile">
          <div class="tile-value">
            <span
              style="color: {responseTimeChange > 0
                ? 'var(--highlight)'
                : 'var(--red)'}">{responseTimeChange >= 0 ? '+' : '-'}{responseTimeChange.toFixed(1)}%</span
            >
          </div>
          <div class="tile-label">Response time</div>
        </div>
    {/if}
  </div>
</div>

<style>
  .card {
    /* width: 100%; */
    flex: 1.2;
    margin: 2em 1em 2em 0;
  }
  .values {
    margin: 30px 30px;
    display: flex;
  }
  .tile-value {
    font-size: 1.4em;
    margin-bottom: 5px;
  }
  .tile-label {
    font-size: 0.8em;
  }
  .tile {
    background: #282828;
    /* border: 1px solid var(--highlight); */
    flex: 1;
    padding: 30px 10px;
    border-radius: 6px;
    margin: 10px;
  }
  @media screen and (max-width: 1580px) {
    .card {
      width: 100%;
    }
  }
  @media screen and (max-width: 1580px) {
    .card {
      margin: 2em 0 2em;
      width: 100%;
    }
  }
</style>
