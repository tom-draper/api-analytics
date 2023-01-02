<script lang="ts">
  import { onMount } from "svelte";
  import periodToDays from "../../lib/period";
  
  function togglePeriod() {
    perHour = !perHour;
  }

  function setPercentageChange(now: number, prev: number) {
    if (prev == 0) {
      percentageChange = null;
    } else {
      percentageChange = (now / prev) * 100 - 100;
    }
  }

  function getUsers(data: RequestsData): Set<string> {
    let users: Set<string> = new Set();
    for (let i = 0; i < data.length; i++) {
      if (data[i].ip_address) {
        users.add(data[i].ip_address)
      }
    }
    return users
  }

  function build() {
    let users = getUsers(data);
    numUsers = users.size;
    
    let prevUsers = getUsers(prevData);
    let prevNumUsers = prevUsers.size;

    setPercentageChange(numUsers, prevNumUsers);

    if (numUsers > 0) {
      let days = periodToDays(period);
      if (days != null) {
        usersPerHour = (numUsers/ (24 * days)).toFixed(2);
      }
    } else {
      usersPerHour = "0";
    }
  }
  
  let numUsers: number = 0;
  let usersPerHour: string;
  let perHour = false;
  let percentageChange: number;
  let mounted = false;
  onMount(() => {
    mounted = true;
  });
  
  $: data && mounted && build();

  export let data: RequestsData, prevData: RequestsData, period: string;
</script>

{#if perHour}
  <button class="card" on:click="{togglePeriod}">
    <div class="card-title">
      Users <span class="per-hour">/ hour</span>
    </div>
    {#if usersPerHour != undefined}
      <div class="value">{usersPerHour}</div>
    {/if}
  </button>
{:else}
  <button class="card" on:click="{togglePeriod}">
    {#if percentageChange != null}
      <div
        class="percentage-change"
        class:positive={percentageChange > 0}
        class:negative={percentageChange < 0}
      >
        ({percentageChange > 0 ? "+" : ""}{percentageChange.toFixed(1)}%)
      </div>
    {/if}
    <div class="card-title">Users</div>
    <div class="value">{numUsers.toLocaleString()}</div>
  </button>
{/if}

<style scoped>
  .card {
    width: calc(200px - 1em);
    margin: 0 2em 0 1em;
    position: relative;
    cursor: pointer;
  }
  .value {
    margin: 20px 0;
    font-size: 1.8em;
    font-weight: 600;
  }
  .percentage-change {
    position: absolute;
    right: 20px;
    top: 20px;
    font-size: 0.8em;
  }
  .positive {
    color: var(--highlight);
  }
  .negative {
    color: rgb(228, 97, 97);
  }

  .per-hour {
    color: var(--dim-text);
    font-size: 0.8em;
    margin-left: 4px;
  }
  button {
    font-size: unset;
    font-family: unset;
  }
</style>
