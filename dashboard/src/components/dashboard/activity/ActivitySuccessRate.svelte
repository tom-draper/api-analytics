<script lang="ts">
  import { onMount } from "svelte";

  function periodToDays(period: string): number {
    if (period == "24-hours") {
      return 1;
    } else if (period == "week") {
      return 8;
    } else if (period == "month") {
      return 30;
    } else if (period == "3-months") {
      return 30 * 3;
    } else if (period == "6-months") {
      return 30 * 6;
    } else if (period == "year") {
      return 365;
    } else {
      return null;
    }
  }

  function daysAgo(date: Date): number {
    let now = new Date();
    return Math.floor((now.getTime() - date.getTime()) / (24 * 60 * 60 * 1000));
  }

  function setSuccessRate() {
    let success = {};
    let minDate = Number.POSITIVE_INFINITY;
    for (let i = 0; i < data.length; i++) {
      let date = new Date(data[i].created_at);
      date.setHours(0, 0, 0, 0);
      // @ts-ignore
      if (!(date in success)) {
        // @ts-ignore
        success[date] = { total: 0, successful: 0 };
      }
      if (data[i].status >= 200 && data[i].status <= 299) {
        // @ts-ignore
        success[date].successful++;
      }
      // @ts-ignore
      success[date].total++;
      if (date as any < minDate) {
        minDate = date as any;
      }
    }

    let days = periodToDays(period);
    if (days == null) {
      days = daysAgo(minDate as any);
    }

    let successArr = new Array(days).fill(-0.1); // -0.1 -> 0
    for (let date in success) {
      let idx = daysAgo(new Date(date));
      successArr[successArr.length-1 - idx] = success[date].successful / success[date].total;
    }

    successRate = successArr;
  }

  function build() {
    setSuccessRate();
  }

  let successRate: any[];
  let mounted = false;
  onMount(() => {
    mounted = true;
  });

  $: successRate;

  $: data && mounted && build();

  export let data: RequestsData, period: string;
</script>

<div class="success-rate-container">
  {#if successRate != undefined}
    <div class="success-rate-title">Success rate</div>
    <div class="errors">
      {#each successRate as value, _}
        <div
          class="error level-{Math.floor(value*10)+1}"
          title="{value >= 0 ? (value * 100).toFixed(1) + '%' : 'No requests'}"
        />
      {/each}
    </div>
  {/if}
</div>

<style>
  .errors {
    display: flex;
    margin-top: 8px;
    margin: 0 10px 0 40px;
  }
  .error {
    background: var(--highlight);
    flex: 1;
    height: 40px;
    margin: 0 0.1%;
    border-radius: 1px;
  }
  .success-rate-container {
    text-align: left;
    font-size: 0.9em;
    color: #707070;
  }
  .success-rate-title {
    margin: 0 0 4px 45px;
  }
  .success-rate-container {
    margin: 1.5em 2.5em 2em;
  }
  .level-0 {
    background: rgb(40, 40, 40);
  }
  .level-1 {
    background: #E46161;
  }
  .level-2 {
    background: #F18359;
  }
  .level-3 {
    background: #F5A65A;
  }
  .level-4 {
    background: #F3C966;
  }
  .level-5 {
    background: #EBEB81;
  }
  .level-6 {
    background: #C7E57D;
  }
  .level-7 {
    background: #A1DF7E;
  }
  .level-8 {
    background: #77D884;
  }
  .level-9 {
    background: #3FCF8E;
  }
</style>