<script lang="ts">

</script>

<div class="card">
  <div class="card-text">
    <div class="url">
      <div class="url-text">
        URL
      </div>
      <input type="text" />
      <button>Add</button>
    </div>
    <div class="detail">Endpoints are pinged by our servers every 30 mins and response <b>status</b> and response <b>time</b> are logged.</div>
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
    background: #1c1c1c;
    border-radius: 4px;
    border: none;
    margin: 0 10px;
    width: 100%;
    padding: 0 5px;
    color: white;
  }
  .url {
    display: flex;
  }
  .url-text {
    margin: auto;
  }
  .detail {
    margin-top: 30px;
    color: #707070;
    font-weight: 500;
    font-size: 0.9em;
  }
  button {
    background: var(--highlight);
    padding: 4px 20px;
    border-radius: 4px;
  }
</style>
