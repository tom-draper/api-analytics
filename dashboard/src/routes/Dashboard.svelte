<script lang="ts">
  import { onMount } from 'svelte';
  import Requests from '../components/dashboard/Requests.svelte';
  import Logo from '../components/dashboard/Logo.svelte';
  import ResponseTimes from '../components/dashboard/ResponseTimes.svelte';
  import Users from '../components/dashboard/Users.svelte';
  import Endpoints from '../components/dashboard/Endpoints.svelte';
  import Footer from '../components/Footer.svelte';
  import SuccessRate from '../components/dashboard/SuccessRate.svelte';
  import Activity from '../components/dashboard/activity/Activity.svelte';
  import Version from '../components/dashboard/Version.svelte';
  import UsageTime from '../components/dashboard/UsageTime.svelte';
  import Location from '../components/dashboard/Location.svelte';
  import Device from '../components/dashboard/device/Device.svelte';
  import { periodToDays } from '../lib/period';
  import genDemoData from '../lib/demo';
  import formatUUID from '../lib/uuid';
  import Settings from '../components/dashboard/Settings.svelte';
  import type { DashboardSettings, Period } from '../lib/settings';
  import { initSettings } from '../lib/settings';
  import type { NotificationState } from '../lib/notification';
  import {
    CREATED_AT,
    PATH,
    STATUS,
    SERVER_URL,
    HOSTNAME,
  } from '../lib/consts';
  import Dropdown from '../components/dashboard/Dropdown.svelte';
  import Notification from '../components/dashboard/Notification.svelte';

  function inPeriod(date: Date, days: number): boolean {
    const periodAgo = new Date();
    periodAgo.setDate(periodAgo.getDate() - days);
    return date > periodAgo;
  }

  function allTimePeriod(_: Date) {
    return true;
  }

  function setPeriodData() {
    const days = periodToDays(settings.period);

    let counted: (date: Date) => boolean = allTimePeriod;
    if (days !== null) {
      counted = (date: Date) => {
        return inPeriod(date, days);
      };
    }

    const dataSubset = [];
    for (let i = 0; i < data.length; i++) {
      if (
        (settings.disable404 && data[i][STATUS] === 404) ||
        (settings.targetEndpoint.path !== null &&
          settings.targetEndpoint.path !== data[i][PATH]) ||
        (settings.targetEndpoint.status !== null &&
          settings.targetEndpoint.status !== data[i][STATUS]) ||
        isHiddenEndpoint(data[i][PATH]) ||
        (settings.hostname !== null && settings.hostname !== data[i][HOSTNAME])
      ) {
        continue;
      }
      const date = data[i][CREATED_AT];
      if (counted(date)) {
        dataSubset.push(data[i]);
      }
    }

    periodData = dataSubset;
  }

  function inPrevPeriod(date: Date, days: number): boolean {
    const startPeriodAgo = new Date();
    startPeriodAgo.setDate(startPeriodAgo.getDate() - days * 2);
    const endPeriodAgo = new Date();
    endPeriodAgo.setDate(endPeriodAgo.getDate() - days);
    return startPeriodAgo < date && date < endPeriodAgo;
  }

  function isHiddenEndpoint(endpoint: string): boolean {
    return (
      settings.hiddenEndpoints.has(endpoint) ||
      (endpoint.charAt(0) === '/' &&
        settings.hiddenEndpoints.has(endpoint.slice(1))) ||
      (endpoint.charAt(endpoint.length - 1) === '/' &&
        settings.hiddenEndpoints.has(endpoint.slice(0, -1))) ||
      //  (endpoint.charAt(0) !== '/' && settings.hiddenEndpoints.has('/' + endpoint)) ||
      (endpoint.charAt(endpoint.length - 1) !== '/' &&
        settings.hiddenEndpoints.has(endpoint + '/')) ||
      wildCardMatch(endpoint)
    );
  }

  function wildCardMatch(endpoint: string): boolean {
    if (endpoint.charAt(endpoint.length - 1) !== '/') {
      endpoint = endpoint + '/';
    }
    for (let value of settings.hiddenEndpoints) {
      if (value.charAt(value.length - 1) === '*') {
        value = value.slice(0, value.length - 1); // Remove asterisk
        // Format both paths with a starting '/' and no trailing '/'
        if (value.charAt(0) !== '/') {
          value = '/' + value;
        }
        if (value.charAt(value.length - 1) !== '/') {
          value = value + '/';
        }
        if (endpoint.slice(0, value.length) === value) {
          return true;
        }
      }
    }
    return false;
  }

  function setPrevPeriodData() {
    const days = periodToDays(settings.period);

    let inPeriod = allTimePeriod;
    if (days !== null) {
      inPeriod = (date) => {
        return inPrevPeriod(date, days);
      };
    }

    const dataSubset = [];
    for (let i = 0; i < data.length; i++) {
      if (
        (settings.disable404 && data[i][STATUS] === 404) ||
        (settings.targetEndpoint.path !== null &&
          settings.targetEndpoint.path !== data[i][PATH]) ||
        (settings.targetEndpoint.status !== null &&
          settings.targetEndpoint.status !== data[i][STATUS]) ||
        isHiddenEndpoint(data[i][PATH]) ||
        (settings.hostname !== null && settings.hostname !== data[i][HOSTNAME])
      ) {
        continue;
      }
      const date = data[i][CREATED_AT];
      if (inPeriod(date)) {
        dataSubset.push(data[i]);
      }
    }
    prevPeriodData = dataSubset;
  }

  function setPeriod(value: Period) {
    settings.period = value;
  }

  type ValueCount = {
    [value: string]: number;
  };

  function sortedFrequencies(freq: ValueCount): string[] {
    const sortedFreq: { value: string; count: number }[] = [];
    for (const value in freq) {
      sortedFreq.push({
        value: value,
        count: freq[value],
      });
    }
    sortedFreq.sort((a, b) => {
      return b.count - a.count;
    });

    const values = [];
    for (const value of sortedFreq) {
      values.push(value.value);
    }

    return values;
  }

  function setHostnames() {
    const hostnameFreq: ValueCount = {};
    for (let i = 0; i < data.length; i++) {
      const hostname = data[i][HOSTNAME];
      if (hostname === null || hostname === '' || hostname === 'null') {
        continue;
      }
      if (!(hostname in hostnameFreq)) {
        hostnameFreq[hostname] = 0;
      }
      hostnameFreq[hostname] += 1;
    }

    const hostnames = sortedFrequencies(hostnameFreq);

    if (hostnames.length > 0) {
      settings.hostname = hostnames[0];
    }
  }

  function toggleEnable404() {
    settings.disable404 = !settings.disable404;
    // Allow button to toggle colour responsively
    setTimeout(() => {
      refreshData;
    }, 10);
  }

  async function fetchData() {
    userID = formatUUID(userID);
    try {
      const response = await fetch(`${SERVER_URL}/api/requests/${userID}`);
      if (response.status === 200) {
        const json = await response.json();
        return json;
      }
    } catch (e) {
      failed = true;
    }
  }

  function parseDates(data: RequestsData) {
    for (let i = 0; i < data.length; i++) {
      data[i][CREATED_AT] = new Date(data[i][CREATED_AT]);
    }
  }

  type PeriodDataCache = {
    [period: string]: {
      periodData: RequestsData;
      prevPeriodData: RequestsData;
    };
  };

  let data: RequestsData;
  let settings: DashboardSettings = initSettings();
  let showSettings: boolean = false;
  let hostnames: string[];
  const notification: NotificationState = {
    message: '',
    style: 'error',
    show: false,
  };
  // let periodDataCache: PeriodDataCache = {};
  let periodData: RequestsData;
  let prevPeriodData: RequestsData;
  const timePeriods: Period[] = [
    '24 hours',
    'Week',
    'Month',
    '6 months',
    'Year',
    'All time',
  ];
  let failed = false;
  let endpointsRendered = false;
  onMount(async () => {
    if (demo) {
      data = genDemoData();
    } else {
      data = await fetchData();
    }
    setPeriod(settings.period);
    setHostnames();
    parseDates(data);

    data?.sort((a, b) => {
      //@ts-ignore
      return a[CREATED_AT] - b[CREATED_AT];
    });
  });

  function refreshData() {
    if (data === undefined) {
      return;
    }

    setPeriodData();
    setPrevPeriodData();
  }

  $: if (settings.targetEndpoint.path == null || settings.targetEndpoint.path) {
    refreshData();
  }

  export let userID: string, demo: boolean;
</script>

{#if periodData}
  <div class="dashboard">
    <div class="button-nav">
      <div class="donate">
        <a
          target="_blank"
          href="https://www.buymeacoffee.com/tomdraper"
          class="donate-link">Donate</a
        >
      </div>
      <button
        class="settings"
        on:click={() => {
          showSettings = true;
        }}
      >
        <img class="settings-icon" src="../img/cog.png" alt="" />
      </button>
      {#if hostnames}
        <Dropdown options={hostnames} bind:selected={settings.hostname} />
      {/if}
      <div class="nav-btn time-period">
        {#each timePeriods as period}
          <button
            class="time-period-btn"
            class:time-period-btn-active={settings.period === period}
            on:click={() => {
              setPeriod(period);
            }}
          >
            {period}
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
            period={settings.period}
          />
          <Users
            data={periodData}
            prevData={prevPeriodData}
            period={settings.period}
          />
        </div>
        <ResponseTimes data={periodData} />
        <Endpoints
          data={periodData}
          bind:targetPath={settings.targetEndpoint.path}
          bind:targetStatus={settings.targetEndpoint.status}
          bind:endpointsRendered
        />
        <Version data={periodData} bind:endpointsRendered />
      </div>
      <div class="right">
        <Activity data={periodData} period={settings.period} />
        <div class="grid-row">
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
<Settings bind:show={showSettings} bind:settings />
<Notification state={notification} />
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
  .time-period {
    display: flex;
    border: 1px solid #2e2e2e;
    border-radius: 4px;
    overflow: hidden;
  }
  .time-period-btn {
    background: var(--background);
    padding: 3px 12px;
    border: none;
    color: var(--dim-text);
    cursor: pointer;
  }
  .time-period-btn-active {
    background: var(--highlight);
    color: black;
  }
  .settings {
    background: transparent;
    outline: none;
    border: none;
    margin-right: 10px;
    cursor: pointer;
    text-align: right;
  }
  .donate {
    margin-left: auto;
    font-weight: 300;
    font-size: 0.85em;
    display: grid;
    place-items: center;
    margin-right: 1em;
  }

  .donate-link {
    color: rgb(73, 73, 73);
    color: rgb(82, 82, 82);
    /* font-family: Arial, 'Noto Sans', */
    /* text-decoration: underline; */
  }
  .settings-icon {
    width: 20px;
    height: 20px;
    filter: contrast(0.5);
    margin-top: 2px;
  }

  @media screen and (max-width: 800px) {
    .donate {
      display: none;
    }
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
