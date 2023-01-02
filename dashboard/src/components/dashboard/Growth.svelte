<script lang="ts">
  import { onMount } from "svelte";

  function periodData(data: RequestsData): any {
    let period = { requests: 0, success: 0, responseTime: 0 };
    for (let i = 0; i < data.length; i++) {
      // @ts-ignore
      period.requests++;
      if (data[i].status >= 200 && data[i].status <= 299) {
        period.success++;
      }
      period.responseTime += data[i].response_time;
    }
    return period;
  }

  function buildWeek() {
    let thisPeriod = periodData(data);
    let lastPeriod = periodData(prevData);

    let requestsChange =
      ((thisPeriod.requests + 1) / (lastPeriod.requests + 1)) * 100 - 100;
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
            class:tile-bad={change.requests > 0}
            class:tile-good={change.requests < 0}
            >{change.requests > 0 ? "+" : ""}{change.requests.toFixed(1)}%</span
          >
        </div>
        <div class="tile-label">Requests</div>
      </div>
      <div class="tile">
        <div class="tile-value">
          <span
            class:tile-bad={change.success > 0}
            class:tile-good={change.success < 0}
            >{change.success > 0 ? "+" : ""}{change.success.toFixed(1)}%</span
          >
        </div>
        <div class="tile-label">Success rate</div>
      </div>
      <div class="tile">
        <div class="tile-value">
          <span
            class:tile-bad={change.responseTime < 0}
            class:tile-good={change.responseTime > 0}
            >{change.responseTime > 0 ? "+" : ""}{change.responseTime.toFixed(
              1
            )}%</span
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
