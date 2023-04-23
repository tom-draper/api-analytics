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
  import Location from "../components/dashboard/Location.svelte";
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
    let days = periodToDays(currentPeriod);

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
    let days = periodToDays(currentPeriod);

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
    currentPeriod = value;
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
        `https://www.apianalytics-server.com/api/requests/${userID}`
      );
      if (response.status === 200) {
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
  let timePeriods = [
    { name: "24-hours", label: "24 hours" },
    { name: "week", label: "Week" },
    { name: "month", label: "Month" },
    { name: "3-month", label: "3 months" },
    { name: "6-month", label: "6 months" },
    { name: "year", label: "Year" },
    { name: "all-time", label: "All time" },
  ];
  let currentPeriod = timePeriods[2].name;
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
        {#each timePeriods as period}
          <button
            class="time-period-btn"
            class:time-period-btn-active={currentPeriod === period.name}
            on:click={() => {
              setPeriod(period.name);
            }}
          >
            {period.label}
          </button>
        {/each}
      </div>
    </div>
    <div class="dashboard-content">
      <div class="left">
        <div class="row">
          <Logo />
          <SuccessRate data={periodData} />
        </div>
        <div class="row">
          <Requests
            data={periodData}
            prevData={prevPeriodData}
            period={currentPeriod}
          />
          <Users
            data={periodData}
            prevData={prevPeriodData}
            period={currentPeriod}
          />
        </div>
        <ResponseTimes data={periodData} />
        <Endpoints data={periodData} />
        <Version data={periodData} />
      </div>
      <div class="right">
        <Activity data={periodData} period={currentPeriod} />
        <div class="grid-row">
          <!-- <Growth data={periodData} prevData={prevPeriodData} /> -->
          <Location data={periodData} />
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
      margin: 1.4em 1em 3.5em;
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
