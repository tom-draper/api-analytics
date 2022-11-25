<script lang="ts">
  import { onMount } from "svelte";

  function pastMonth(date: Date): boolean {
    let monthAgo = new Date();
    monthAgo.setDate(monthAgo.getDate() - 30);
    return date > monthAgo;
  }



  function build() {
    let successRate = {}
    for (let i = 0; i < data.length; i++) {
        let date = new Date(data[i].created_at);
        if (pastMonth(date)) {
            date.setHours(0, 0, 0, 0)
            if (!(date in successRate)) {
                successRate[date] = {total: 0, successful: 0}
            }
            if (data[i].status >= 200 && data[i].status <= 299) {
                successRate[date].successful++
            }
            successRate[date].total++
        }
    }
  }

  onMount(() => {
    build();
  });

  export let data: any;
</script>

<div class="card">
  <div class="card-title">Past Month</div>

  <div class="errors">
    {#each Array(60) as x, i}
      <div class="error" />
    {/each}
  </div>
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
