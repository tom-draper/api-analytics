<script lang="ts">
  import { onMount } from "svelte";

  type Period = {
    requests: number,
    users: number,
    success: number,
    responseTime: number
  }

  function periodData(data: RequestsData): Period {
    let period = { requests: data.length, users: 0, success: 0, responseTime: 0 };
    let users: Set<string> = new Set();
    for (let i = 0; i < data.length; i++) {
      // @ts-ignore
      if (data[i].status >= 200 && data[i].status <= 299) {
        period.success++;
      }
      period.responseTime += data[i].response_time;
      if (data[i].ip_address != "" && data[i].ip_address != null) {
        users.add(data[i].ip_address);
      }
    }
    period.users = users.size;
    return period;
  }

  function buildWeek() {
    let thisPeriod = periodData(data);
    let lastPeriod = periodData(prevData);

    let requestsChange =
      ((thisPeriod.requests + 1) / (lastPeriod.requests + 1)) * 100 - 100;
    let usersChange = 
      ((thisPeriod.users + 1) / (lastPeriod.users + 1)) * 100 - 100;
    let successChange =
      (((thisPeriod.success + 1) / (thisPeriod.requests + 1) + 1) /
        ((lastPeriod.success + 1) / (lastPeriod.requests + 1) + 1)) *
        100 -
      100;
    let responseTimeChange =
      (((thisPeriod.responseTime + 1) / (thisPeriod.requests + 1) + 1) /
        ((lastPeriod.responseTime + 1) / (lastPeriod.requests + 1) + 1)) *
        100 -
      100;
    change = {
      requests: requestsChange,
      users: usersChange,
      success: successChange,
      responseTime: responseTimeChange,
    };
  }

  let change: any;
  let mounted = false;
  onMount(() => {
    mounted = true;
  });

  $: data && mounted && buildWeek();

  export let data: RequestsData, prevData: RequestsData;
</script>

<div class="card">
  <div class="card-title">Growth</div>
  <div class="values">
    {#if change != undefined}
      <div class="tile">
        <div class="tile-value">
          <span
            class:tile-bad={change.requests < 0}
            class:tile-good={change.requests > 0}
            >
            {#if change.requests > 0}
              <img class="arrow" src="../img/up.png" alt="" />
            {:else if change.requests < 0}
              <img class="arrow" src="../img/down.png" alt="" />
            {/if}
            {Math.abs(change.requests).toFixed(1)}%</span
          >
        </div>
        <div class="tile-label">Requests</div>
      </div>
      <div class="tile">
        <div class="tile-value">
          <span
            class:tile-bad={change.users < 0}
            class:tile-good={change.users > 0}
            >
            {#if change.users > 0}
              <img class="arrow" src="../img/up.png" alt="" />
            {:else if change.users < 0}
              <img class="arrow" src="../img/down.png" alt="" />
            {/if}
            {Math.abs(change.users).toFixed(1)}%</span
            >
          </div>
          <div class="tile-label">Users</div>
        </div>
        <div class="tile">
          <div class="tile-value">
            <span
            class:tile-bad={change.success < 0}
            class:tile-good={change.success > 0}
            >
            {#if change.success > 0}
              <img class="arrow" src="../img/up.png" alt="" />
            {:else if change.success < 0}
              <img class="arrow" src="../img/down.png" alt="" />
            {/if}
            {Math.abs(change.success).toFixed(1)}%</span
          >
        </div>
        <div class="tile-label">Success rate</div>
      </div>
      <div class="tile">
        <div class="tile-value">
          <span
          class:tile-bad={change.responseTime > 0}
          class:tile-good={change.responseTime < 0}
          >
            <!-- Response time -- down is good -->
            {#if change.responseTime < 0}
              <img class="arrow" src="../img/good-down.png" alt="" />
            {:else if change.responseTime > 0}
              <img class="arrow" src="../img/bad-up.png" alt="" />
            {/if}
            {Math.abs(change.responseTime).toFixed(1)}%</span
          >
        </div>
        <div class="tile-label">Response time</div>
      </div>
    {/if}
  </div>
</div>

<style scoped>
  .card {
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
    font-weight: 600;
  }
  .tile-label {
    font-size: 0.8em;
  }
  .tile {
    background: #282828;
    flex: 1;
    padding: 30px 10px;
    border-radius: 6px;
    margin: 10px;
  }
  .tile-bad {
    color: var(--red);
  }
  .tile-good {
    color: var(--highlight);
  }
  .arrow {
    height: 15px;
  }
  @media screen and (max-width: 1600px) {
    .card {
      width: 100%;
      margin: 2em 0 2em;
    }
  }
  @media screen and (max-width: 650px) {
    .values {
      flex-direction: column;
    }
  }
</style>
