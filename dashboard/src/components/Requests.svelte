<script lang="ts">
  export let data: RequestsData;
</script>

<div class="card" title="Total">
  <div class="card-title">
    Requests
  </div>
  <div class="value">{data.length.toLocaleString()}</div>
</div>

<style>
  .card {
    width: calc(200px - 1em);
    margin: 0 1em 0 2em;
  }
  .value {
    margin: 20px 0;
    font-size: 1.8em;
    font-weight: 600;
  }
</style>
