<script lang="ts">
  import { onMount } from "svelte";

  function pastMonth(date: Date): boolean {
    let monthAgo = new Date();
    monthAgo.setDate(monthAgo.getDate() - 30);
    return date > monthAgo;
  }

  let colors = [
    'grey',
    '#3FCF8E',
    '#77D884',
    '#A1DF7E',
    '#C7E57D',
    '#EBEB81',
    '#EBEB81',
    '#F3C966',
    '#F5A65A',
    '#F18359',
    '#E46161'
  ]

  function daysAgo(date: Date): number {
    let now = new Date()
    return Math.floor((now.getTime() - date.getTime())/(24 * 60 * 60 * 1000))
  }

  function build() {
    let successRate = {}
    for (let i = 0; i < data.length; i++) {
        let date = new Date(data[i].created_at);
        if (pastMonth(date)) {
            date.setHours(0, 0, 0, 0)
            // @ts-ignore
            if (!(date in successRate)) {
                // @ts-ignore
                successRate[date] = {total: 0, successful: 0}
            }
            if (data[i].status >= 200 && data[i].status <= 299) {
                // @ts-ignore
                successRate[date].successful++
            }
            // @ts-ignore
            successRate[date].total++
        }
    }

    console.log(successRate)

    let sr = new Array(60).fill(0)

    for (let date in successRate) {
        let idx = daysAgo(new Date(date))
        console.log(date, idx)
        sr[sr.length-idx] = (successRate[date].successful / successRate[date].total)
    }
    console.log(sr)
    _successRate = sr;
  }

  let _successRate
  onMount(() => {
    build();
  });

  export let data: any;
</script>

<div class="card">
  <div class="card-title">Past Month</div>

  {#if _successRate != undefined}
    <div class="errors">
        {#each _successRate as value, i}
        <div class="error" style="background: {colors[Math.floor(value*10)]}" title="{(value*100).toFixed(1)}%"/>
        {/each}
    </div>
  {/if}
</div>

<style>
  .card {
    margin: 0;
    width: 100%;
  }
  .errors {
    display: flex;
    margin: 2em;
  }
  .error {
    background: var(--highlight);
    flex: 1;
    height: 40px;
    margin: 0 1px;
    border-radius: 1px;
  }
</style>
