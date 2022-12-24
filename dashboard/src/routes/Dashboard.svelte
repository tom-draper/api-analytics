<script lang="ts">
  import { onMount } from "svelte";
  import Requests from "../components/dashboard/Requests.svelte";
  import Welcome from "../components/dashboard/Welcome.svelte";
  import RequestsPerHour from "../components/dashboard/RequestsPerHour.svelte";
  import ResponseTimes from "../components/dashboard/ResponseTimes.svelte";
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

  function formatUUID(userID: string): string {
    return `${userID.slice(0, 8)}-${userID.slice(8, 12)}-${userID.slice(
      12,
      16
    )}-${userID.slice(16, 20)}-${userID.slice(20)}`;
  }

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
    for (let i = 0; i < data.length; i++) {
      let date = new Date(data[i].created_at);
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
    for (let i = 0; i < data.length; i++) {
      let date = new Date(data[i].created_at);
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
  async function fetchData() {
    userID = formatUUID(userID);
    // Fetch page ID
    try {
      const response = await fetch(
        `https://api-analytics-server.vercel.app/api/user-data/${userID}`
      );
      if (response.status == 200) {
        const json = await response.json();
        data = json.value;
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
    <div class="time-period">
      <button
        class="time-period-btn {period == '24-hours'
          ? 'time-period-btn-active'
          : ''}"
        on:click={() => {
          setPeriod("24-hours");
        }}
      >
        24 hours
      </button>
      <button
        class="time-period-btn {period == 'week'
          ? 'time-period-btn-active'
          : ''}"
        on:click={() => {
          setPeriod("week");
        }}
      >
        Week
      </button>
      <button
        class="time-period-btn {period == 'month'
          ? 'time-period-btn-active'
          : ''}"
        on:click={() => {
          setPeriod("month");
        }}
      >
        Month
      </button>
      <button
        class="time-period-btn {period == '3-months'
          ? 'time-period-btn-active'
          : ''}"
        on:click={() => {
          setPeriod("3-months");
        }}
      >
        3 months
      </button>
      <button
        class="time-period-btn {period == '6-months'
          ? 'time-period-btn-active'
          : ''}"
        on:click={() => {
          setPeriod("6-months");
        }}
      >
        6 months
      </button>
      <button
        class="time-period-btn {period == 'year'
          ? 'time-period-btn-active'
          : ''}"
        on:click={() => {
          setPeriod("year");
        }}
      >
        Year
      </button>
      <button
        class="time-period-btn {period == 'all-time'
          ? 'time-period-btn-active'
          : ''}"
        on:click={() => {
          setPeriod("all-time");
        }}
      >
        All time
      </button>
    </div>
    <div class="left">
      <div class="row">
        <Welcome />
        <SuccessRate data={periodData} />
      </div>
      <div class="row">
        <Requests data={periodData} prevData={prevPeriodData} {period} />
        <RequestsPerHour data={periodData} {period} />
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
{:else if failed}
  <div class="no-requests">No requests currently logged.</div>
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
    margin: 5em;
    display: flex;
    position: relative;
  }
  .row {
    display: flex;
  }
  .grid-row {
    display: flex;
  }
  .right {
    flex-grow: 1;
    margin-right: 2em;
  }
  .no-requests {
    height: 70vh;
    font-size: 1.5em;
    display: grid;
    place-items: center;
    color: var(--highlight);
  }
  .placeholder {
    min-height: 85vh;
    display: grid;
    place-items: center;
  }
  .time-period {
    position: absolute;
    display: flex;
    right: 2em;
    top: -2.2em;
    border: 1px solid #2e2e2e;
    border-radius: 4px;
    overflow: hidden;
  }
  .time-period-btn {
    background: #232323;
    padding: 3px 12px;
    border: none;
    color: #707070;
    cursor: pointer;
  }
  .time-period-btn-active {
    background: var(--highlight);
    color: black;
  }
  @media screen and (max-width: 1580px) {
    .grid-row {
      flex-direction: column;
    }
  }
  @media screen and (max-width: 1100px) {
    .dashboard {
      margin: 2em 0;
    }
  }
</style>
