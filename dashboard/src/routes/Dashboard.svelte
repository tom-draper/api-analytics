<script lang="ts">
  import { onMount } from "svelte";
  import Requests from "../components/dashboard/Requests.svelte";
  import Logo from "../components/dashboard/Logo.svelte";
  import ResponseTimes from "../components/dashboard/ResponseTimes.svelte";
  import Users from "../components/dashboard/Users.svelte";
  import Endpoints from "../components/dashboard/Endpoints.svelte";
  import Footer from "../components/Footer.svelte";
  import SuccessRate from "../components/dashboard/SuccessRate.svelte";
  import Activity from "../components/dashboard/activity/Activity.svelte";
  import Version from "../components/dashboard/Version.svelte";
  import UsageTime from "../components/dashboard/UsageTime.svelte";
  import Growth from "../components/dashboard/Growth.svelte";
  import Device from "../components/dashboard/device/Device.svelte";
  import periodToDays from "../lib/period";
  import genDemoData from "../lib/demo";
  import formatUUID from "../lib/uuid";

  function inPeriod(date: Date, days: number): boolean {
    let periodAgo = new Date();
    periodAgo.setDate(periodAgo.getDate() - days);
    return date > periodAgo;
  }

  function allTimePeriod(date: Date) {
    return true;
  }

  function setPeriodData() {
    let days = periodToDays(period);

    let counted = allTimePeriod;
    if (days != null) {
      counted = (date) => {
        return inPeriod(date, days);
      };
    }

    let dataSubset = [];
    for (let i = 1; i < data.length; i++) {
      if (disable404 && data[i][5] == 404) {
        continue;
      }
      let date = new Date(data[i][7]);
      if (counted(date)) {
        dataSubset.push(data[i]);
      }
    }

    periodData = dataSubset;
  }

  function inPrevPeriod(date: Date, days: number): boolean {
    let startPeriodAgo = new Date();
    startPeriodAgo.setDate(startPeriodAgo.getDate() - days * 2);
    let endPeriodAgo = new Date();
    endPeriodAgo.setDate(endPeriodAgo.getDate() - days);
    return startPeriodAgo < date && date < endPeriodAgo;
  }

  function setPrevPeriodData() {
    let days = periodToDays(period);

    let inPeriod = allTimePeriod;
    if (days != null) {
      inPeriod = (date) => {
        return inPrevPeriod(date, days);
      };
    }

    let dataSubset = [];
    for (let i = 1; i < data.length; i++) {
      if (disable404 && data[i][5] == 404) {
        continue;
      }
      let date = new Date(data[i][7]);
      if (inPeriod(date)) {
        dataSubset.push(data[i]);
      }
    }
    prevPeriodData = dataSubset;
  }

  function setPeriod(value: string) {
    period = value;
    setPeriodData();
    setPrevPeriodData();
  }

  function toggleEnable404() {
    disable404 = !disable404;
    // Allow button to toggle colour responsively
    setTimeout(() => {
      setPeriodData();
      setPrevPeriodData();
    }, 10);
  }

  async function fetchData() {
    userID = formatUUID(userID);
    try {
      const response = await fetch(
        // `https://api-analytics-server.vercel.app/api/requests/${userID}`
        `http://213.168.248.206/api/requests/${userID}`
      );
      if (response.status == 200) {
        const json = await response.json();
        data = json;
        console.log(data);
        setPeriod("month");
      }
    } catch (e) {
      failed = true;
    }
  }

  let data: RequestsData;
  let periodData: RequestsData;
  let prevPeriodData: RequestsData;
  let period = "month";
  let failed = false;
  let disable404 = false;
  onMount(() => {
    if (demo) {
      data = genDemoData() as RequestsData;
      setPeriod("month");
    } else {
      fetchData();
    }
  });

  export let userID: string, demo: boolean;
</script>

{#if periodData != undefined}
  <div class="dashboard">
    <div class="button-nav">
      <div class="nav-btn enable-404">
        <button
          class="enable-404-btn"
          on:click={toggleEnable404}
          class:time-period-btn-active={disable404}>Disable 404</button
        >
      </div>
      <div class="nav-btn time-period">
        <button
          class="time-period-btn"
          class:time-period-btn-active={period === "24-hours"}
          on:click={() => {
            setPeriod("24-hours");
          }}
        >
          24 hours
        </button>
        <button
          class="time-period-btn"
          class:time-period-btn-active={period === "week"}
          on:click={() => {
            setPeriod("week");
          }}
        >
          Week
        </button>
        <button
          class="time-period-btn"
          class:time-period-btn-active={period === "month"}
          on:click={() => {
            setPeriod("month");
          }}
        >
          Month
        </button>
        <button
          class="time-period-btn"
          class:time-period-btn-active={period === "3-months"}
          on:click={() => {
            setPeriod("3-months");
          }}
        >
          3 months
        </button>
        <button
          class="time-period-btn"
          class:time-period-btn-active={period === "6-months"}
          on:click={() => {
            setPeriod("6-months");
          }}
        >
          6 months
        </button>
        <button
          class="time-period-btn"
          class:time-period-btn-active={period === "year"}
          on:click={() => {
            setPeriod("year");
          }}
        >
          Year
        </button>
        <button
          class="time-period-btn"
          class:time-period-btn-active={period === "all-time"}
          on:click={() => {
            setPeriod("all-time");
          }}
        >
          All time
        </button>
      </div>
    </div>
    <div class="dashboard-content">
      <div class="left">
        <div class="row">
          <Logo />
          <SuccessRate data={periodData} />
        </div>
        <div class="row">
          <Requests data={periodData} prevData={prevPeriodData} {period} />
          <Users data={periodData} prevData={prevPeriodData} {period} />
        </div>
        <ResponseTimes data={periodData} />
        <Endpoints data={periodData} />
        <Version data={periodData} />
      </div>
      <div class="right">
        <Activity data={periodData} {period} />
        <div class="grid-row">
          <Growth data={periodData} prevData={prevPeriodData} />
          <Device data={periodData} />
        </div>
        <UsageTime data={periodData} />
      </div>
    </div>
  </div>
{:else if failed}
  <img class="no-requests" src="../img/no-requests-logged.png" alt="" />
{:else}
  <div class="placeholder" style="min-height: 85vh;">
    <div class="spinner">
      <div class="loader" />
    </div>
  </div>
{/if}
<Footer />

<style scoped>
  .dashboard {
    min-height: 90vh;
  }
  .dashboard {
    margin: 1.4em 5em 5em;
  }
  .dashboard-content {
    margin-top: 1.4em;
    display: flex;
    position: relative;
  }
  .row {
    display: flex;
    margin-bottom: 2em;
  }
  .grid-row {
    display: flex;
  }
  .no-requests {
    width: 350px;
    margin: 20vh 0;
  }
  .left {
    margin: 0 2em;
  }
  .right {
    flex-grow: 1;
    margin-right: 2em;
  }
  .placeholder {
    min-height: 85vh;
    display: grid;
    place-items: center;
  }
  .button-nav {
    margin: 2.5em 2em 0;
    display: flex;
  }
  .nav-btn {
    margin-left: auto;
  }
  .enable-404 {
    text-align: right;
    flex-grow: 1;
    margin-right: 1.5em;
    width: fit-content;
  }
  .time-period {
    display: flex;
    border: 1px solid #2e2e2e;
    border-radius: 4px;
    overflow: hidden;
  }
  .enable-404-btn,
  .time-period-btn {
    background: var(--light-background);
    padding: 3px 12px;
    border: none;
    color: var(--dim-text);
    cursor: pointer;
  }
  .enable-404-btn {
    border: 1px solid #2e2e2e;
    padding: 3px 12px;
    border-radius: 4px;
  }
  .time-period-btn-active {
    background: var(--highlight);
    color: black;
  }
  @media screen and (max-width: 1600px) {
    .grid-row {
      flex-direction: column;
    }
  }
  @media screen and (max-width: 1300px) {
    .dashboard {
      margin: 0;
    }
    .dashboard-content {
      margin: 1.4em 1em 1em;
    }
    .button-nav {
      margin: 2.5em 3em 0;
    }
  }
  @media screen and (max-width: 1030px) {
    .dashboard-content {
      flex-direction: column;
    }
    .right,
    .left {
      margin: 0 2em;
    }
  }
  @media screen and (max-width: 750px) {
    .button-nav {
      flex-direction: column;
    }
    .enable-404 {
      margin: 0 0 1em auto;
    }
  }
  @media screen and (max-width: 600px) {
    .right,
    .left {
      margin: 0 1em;
    }
    .time-period {
      right: 1em;
    }
  }
</style>
