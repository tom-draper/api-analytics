<script lang="ts">
  import { onMount } from "svelte";

  function pastMonth(date: Date): boolean {
    let monthAgo = new Date();
    monthAgo.setDate(monthAgo.getDate() - 60);
    return date > monthAgo;
  }

  let colors = [
    "#444444", // Grey (no requests)
    "#E46161", // Red
    "#F18359",
    "#F5A65A",
    "#F3C966",
    "#EBEB81", // Yellow
    "#C7E57D",
    "#A1DF7E",
    "#77D884",
    "#3FCF8E", // Green
  ];

  function daysAgo(date: Date): number {
    let now = new Date();
    return Math.floor((now.getTime() - date.getTime()) / (24 * 60 * 60 * 1000));
  }

  function setSuccessRate() {
    let success = {};
    for (let i = 0; i < data.length; i++) {
      let date = new Date(data[i].created_at);
      if (pastMonth(date)) {
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
      }
    }

    let successArr = new Array(60).fill(-0.1); // -0.1 -> 0
    for (let date in success) {
      let idx = daysAgo(new Date(date));
      successArr[successArr.length - idx] =
        success[date].successful / success[date].total;
    }
    successRate = successArr;
  }

  function build() {
    setSuccessRate();
  }

  let successRate: any[];
  onMount(() => {
    build();
  });

  export let data: RequestsData;
</script>

<div class="success-rate-container">
  {#if successRate != undefined}
    <div class="success-rate-title">Success rate</div>
    <div class="errors">
      {#each successRate as value, _}
        <div
          class="error"
          style="background: {colors[Math.floor(value * 10) + 1]}"
          title="{(value * 100).toFixed(1)}%"
        />
      {/each}
    </div>
  {/if}
</div>

<style>
  .errors {
    display: flex;
    margin-top: 8px;
    margin: 0 15px 0 40px;
  }
  .error {
    background: var(--highlight);
    flex: 1;
    height: 40px;
    margin: 0 1px;
    border-radius: 1px;
  }
  .success-rate-container {
    text-align: left;
    font-size: 0.9em;
    color: #707070;
    /* color: rgb(68, 68, 68); */
  }
  .success-rate-title {
    margin: 0 0 3px 45px;
  }
  .success-rate-container {
    margin: 1.5em 2em 2em;
  }
</style>
