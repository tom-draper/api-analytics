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

  function formatUUID(userID: string): string {
    return `${userID.slice(0, 8)}-${userID.slice(8, 12)}-${userID.slice(
      12,
      16
    )}-${userID.slice(16, 20)}-${userID.slice(20)}`;
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

  function inPeriod(date: Date, days: number): boolean {
    let periodAgo = new Date();
    periodAgo.setDate(periodAgo.getDate() - days);
    return date > periodAgo;
  }

  function allTimePeriod(date: Date) {
    return true;
  }

  function periodToDays(period: string): number {
    if (period == "24-hours") {
      return 1;
    } else if (period == "week") {
      return 8;
    } else if (period == "month") {
      return 30;
    } else if (period == "3-months") {
      return 30 * 3;
    } else if (period == "6-months") {
      return 30 * 6;
    } else if (period == "year") {
      return 365;
    } else {
      return null;
    }
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

    let count = 0;
    for (let i = 0; i < dataSubset.length; i++) {
      if (dataSubset[i].status >= 400 && dataSubset[i].status < 500) {
        count++;
      }
    }
    console.log(count);
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

  function getDemoStatus(date: Date, status: number) {
    let start = new Date();
    start.setDate(start.getDate() - 100);
    let end = new Date();
    end.setDate(end.getDate() - 96);

    if (date > start && date < end) {
      return 400;
    } else {
      return status;
    }
  }

  function getDemoUserAgent() {
    let p = Math.random();
    if (p < 0.19) {
      return "Mozilla/5.0 (iPhone; CPU iPhone OS 13_5_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.1.1 Mobile/15E148 Safari/604.1";
    } else if (p < 0.3) {
      return "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0) Gecko/20100101 Firefox/47.0";
    } else if (p < 0.34) {
      return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59";
    } else if (p < 0.36) {
      return "curl/7.64.1";
    } else if (p < 0.39) {
      return "PostmanRuntime/7.26.5";
    } else if (p < 0.4) {
      return "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.106 Safari/537.36 OPR/38.0.2220.41";
    } else if (p < 0.4) {
      return "Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:42.0) Gecko/20100101 Firefox/42.0";
    } else {
      return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36";
    }
  }

  function addDemoSamples(
    demoData: any[],
    endpoint: string,
    status: number,
    count: number,
    maxDaysAgo: number,
    minDaysAgo: number,
    maxResponseTime: number,
    minResponseTime: number
  ) {
    for (let i = 0; i < count; i++) {
      let date = new Date();
      date.setDate(
        date.getDate() - Math.floor(Math.random() * maxDaysAgo + minDaysAgo)
      );
      date.setHours(Math.floor(Math.random() * 24));
      demoData.push({
        hostname: "demo-api.com",
        path: endpoint,
        user_agent: getDemoUserAgent(),
        method: 0,
        status: getDemoStatus(date, status),
        response_time: Math.floor(
          Math.random() * maxResponseTime + minResponseTime
        ),
        created_at: date.toISOString(),
      });
    }
  }

  function genDemoData() {
    let demoData = [];

    addDemoSamples(demoData, "/v1/", 200, 18000, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v1/", 400, 1000, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v1/account", 200, 8000, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v1/account", 400, 1200, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v1/help", 200, 700, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v1/help", 400, 70, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v2/", 200, 35000, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v2/", 400, 200, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v2/account", 200, 14000, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v2/account", 400, 3000, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v2/account/update", 200, 6000, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v2/account/update", 400, 400, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v2/help", 200, 6000, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v2/help", 400, 400, 650, 0, 240, 55);

    addDemoSamples(demoData, "/v2/account", 200, 16000, 450, 0, 100, 30);
    addDemoSamples(demoData, "/v2/account", 400, 2000, 450, 0, 100, 30);

    addDemoSamples(demoData, "/v2/help", 200, 8000, 300, 0, 100, 30);
    addDemoSamples(demoData, "/v2/help", 400, 800, 300, 0, 100, 30);

    addDemoSamples(demoData, "/v2/", 200, 4000, 200, 0, 100, 30);
    addDemoSamples(demoData, "/v2/", 400, 800, 200, 0, 100, 30);

    addDemoSamples(demoData, "/v2/", 200, 3000, 100, 0, 100, 30);
    addDemoSamples(demoData, "/v2/", 400, 500, 100, 0, 100, 30);

    addDemoSamples(demoData, "/v2/", 200, 1000, 60, 0, 100, 30);
    addDemoSamples(demoData, "/v2/", 400, 50, 60, 0, 100, 30);

    addDemoSamples(demoData, "/v2/", 200, 500, 40, 0, 100, 30);
    addDemoSamples(demoData, "/v2/", 400, 25, 40, 0, 100, 30);

    addDemoSamples(demoData, "/v2/", 200, 250, 10, 0, 100, 30);
    addDemoSamples(demoData, "/v2/", 400, 20, 10, 0, 100, 30);

    addDemoSamples(demoData, "/v2/", 200, 125, 5, 0, 100, 30);
    addDemoSamples(demoData, "/v2/", 400, 10, 5, 0, 100, 30);

    data = demoData;
    setPeriod("month");
  }

  let data: RequestsData;
  let periodData: RequestsData;
  let prevPeriodData: RequestsData;
  let period = "month";
  let failed = false;
  onMount(() => {
    if (demo) {
      genDemoData();
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
        <Requests data={periodData} prevData={prevPeriodData} />
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

<style>
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
