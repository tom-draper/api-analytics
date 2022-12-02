<script lang="ts">
  import PastMonthRequests from "./PastMonthRequests.svelte";
  import PastMonthResponseTime from "./PastMonthResponseTime.svelte";
  import PastMonthSuccessRate from "./PastMonthSuccessRate.svelte";

  export let data: RequestsData;
</script>

<div class="card">
  <div class="card-title">Past Month</div>
  <PastMonthRequests {data} />
  <PastMonthResponseTime {data} />
  <PastMonthSuccessRate {data} />
</div>

<style>
  .card {
    margin: 0;
    width: 100%;
  }

</style>
