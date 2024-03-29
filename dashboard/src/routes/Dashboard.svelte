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
  import Dropdown from '../components/dashboard/Dropdown.svelte';
  import Notification from '../components/dashboard/Notification.svelte';
  import exportCSV from '../lib/exportData';
  import { ColumnIndex, columns, serverURL } from '../lib/consts';
  import Error from '../components/dashboard/Error.svelte';

  function inPeriod(date: Date, days: number): boolean {
    const periodAgo = new Date();
    periodAgo.setDate(periodAgo.getDate() - days);
    return date > periodAgo;
  }

  function allTimePeriod(_: Date) {
    return true;
  }

  function setData() {
    const days = periodToDays(settings.period);

    let inRange: (date: Date) => boolean;
    if (days !== null) {
      inRange = (date: Date) => {
        return inPeriod(date, days);
      };
    } else {
      // If period is null, set to all time
      inRange = allTimePeriod;
    }

    let inPrevRange: (date: Date) => boolean;
    if (days !== null) {
      inPrevRange = (date) => {
        return inPrevPeriod(date, days);
      };
    } else {
      // If period is null, set to all time
      inPrevRange = allTimePeriod;
    }

    const dataSubset = [];
    const prevDataSubset = [];
    for (let i = 0; i < data.length; i++) {
      // Created inverted version of the if statement to reduce nesting
      const request = data[i];
      const status = request[ColumnIndex.Status];
      const path = request[ColumnIndex.Path];
      const hostname = request[ColumnIndex.Hostname];
      const location = request[ColumnIndex.Location];
      if (
        (!settings.disable404 || status !== 404) &&
        (settings.targetEndpoint.path === null ||
          settings.targetEndpoint.path === path) &&
        (settings.targetEndpoint.status === null ||
          settings.targetEndpoint.status === status) &&
        (settings.targetLocation === null ||
          settings.targetLocation === location) &&
        !isHiddenEndpoint(path) &&
        (settings.hostname === null || settings.hostname === hostname)
      ) {
        const date = request[ColumnIndex.CreatedAt];
        if (inRange(date)) {
          dataSubset.push(request);
        } else if (inPrevRange(date)) {
          prevDataSubset.push(request);
        }
      }
    }

    periodData = dataSubset;
    prevPeriodData = prevDataSubset;
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
      (endpoint.charAt(0) !== '/' &&
        settings.hiddenEndpoints.has('/' + endpoint)) ||
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

  function setPeriod(value: Period) {
    settings.period = value;
  }

  type ValueCount = {
    [value: string]: number;
  };

  function sortedFrequencies(freq: ValueCount): string[] {
    return Object.keys(freq)
      .map((value) => {
        return {
          value: value,
          count: freq[value],
        };
      })
      .sort((a, b) => {
        return b.count - a.count;
      })
      .map((value) => value.value);
  }

  function getHostnames() {
    const hostnameFreq: ValueCount = {};
    for (let i = 0; i < data.length; i++) {
      const hostname = data[i][ColumnIndex.Hostname];
      if (hostname === null || hostname === '' || hostname === 'null') {
        continue;
      }
      hostnameFreq[hostname] |= 0;
      hostnameFreq[hostname] += 1;
    }

    return sortedFrequencies(hostnameFreq);
  }

  function setHostnames() {
    hostnames = getHostnames();
  }

  async function fetchData(): Promise<DashboardData> {
    userID = formatUUID(userID);
    try {
      const response = await fetch(`${serverURL}/api/requests/${userID}`);
      if (response.status === 200) {
        return await response.json();
      } else {
        fetchFailed = true;
      }
    } catch (e) {
      fetchFailed = true;
    }
  }

  function parseDates(data: RequestsData) {
    for (let i = 0; i < data.length; i++) {
      data[i][ColumnIndex.CreatedAt] = new Date(data[i][ColumnIndex.CreatedAt]);
    }
  }

  let data: RequestsData;
  let userAgents: UserAgents;
  let settings: DashboardSettings = initSettings();
  let showSettings: boolean = false;
  let hostnames: string[];
  const notification: NotificationState = {
    message: '',
    style: 'error',
    show: false,
  };
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
  let fetchFailed = false;
  let endpointsRendered = false;
  onMount(async () => {
    const dashboardData = await getDashboardData();
    data = dashboardData.requests;
    userAgents = dashboardData.user_agents;

    setPeriod(settings.period);
    setHostnames();
    parseDates(data);

    data?.sort((a, b) => {
      return (
        a[ColumnIndex.CreatedAt].getTime() - b[ColumnIndex.CreatedAt].getTime()
      );
    });

    console.log(data);
  });

  async function getDashboardData() {
    if (demo) {
      return genDemoData();
    }
    return await fetchData();
  }

  function getUserAgent(id: number): string {
    if (id in userAgents) {
      return userAgents[id];
    }
    return '';
  }

  function refreshData() {
    if (data === undefined) {
      return;
    }

    setData();
  }

  // If target path/location changes or is reset, refresh data with this filter change
  $: if (
    settings.targetEndpoint.path === null ||
    settings.targetEndpoint.path ||
    settings.targetLocation === null ||
    settings.targetLocation
  ) {
    refreshData();
  }

  export let userID: string, demo: boolean;
</script>

{#if periodData && data.length > 0}
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
        <div class="dropdown-container">
          <Dropdown options={hostnames.slice(0, 25)} bind:selected={settings.hostname} defaultOption={'All hostnames'} />
        </div>
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
          <Location
            data={periodData}
            bind:targetLocation={settings.targetLocation}
          />
          <Device data={periodData} {getUserAgent} />
        </div>
        <UsageTime data={periodData} />
      </div>
    </div>
  </div>
{:else if periodData && data.length <= 0}
  <Error reason={'no-requests'} />
{:else if fetchFailed}
  <Error reason={'error'} />
{:else}
  <div class="placeholder">
    <div class="spinner">
      <div class="loader" />
    </div>
  </div>
{/if}
<Settings
  bind:show={showSettings}
  bind:settings
  exportCSV={() => {
    exportCSV(periodData, columns, userAgents);
  }}
/>
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
  .left {
    margin: 0 2em;
  }
  .right {
    flex-grow: 1;
    margin-right: 2em;
  }
  .placeholder {
    min-height: 82vh;
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
    height: 27px;
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

  .dropdown-container {
    margin-right: 10px;
  }

  .donate-link {
    color: rgb(73, 73, 73);
    color: rgb(82, 82, 82);
    color: #464646;
    /* font-family: Arial, 'Noto Sans', */
    /* text-decoration: underline; */
    transition: 0.1s;
  }
  .donate-link:hover {
    color: var(--highlight);
  }
  .settings-icon {
    width: 20px;
    height: 20px;
    filter: contrast(0.45);
    margin-top: 2px;
    transition: 0.1s;
  }
  .settings-icon:hover {
    filter: contrast(0.01);
  }

  @media screen and (max-width: 800px) {
    .donate {
      display: none;
    }
    .settings {
      margin-left: auto;
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
  @media screen and (max-width: 800px) {
    .button-nav {
      flex-direction: column;
    }
    .dropdown-container {
      margin-left: auto;
      margin-right: 0;
      margin: -30px 0 0 auto;
    }
    .time-period {
      margin-top: 15px;
    }
    .time-period-btn {
      flex: 1;
    }
    .settings {
      margin-left: 0;
      margin-right: auto;
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
    .button-nav {
      margin: 2.5em 2em 0;
    }
    .time-period-btn {
      padding: 3px 0;
    }
  }
  @media screen and (max-width: 500px) {
    .time-period-btn {
      flex-grow: 1;
      flex: auto;
    }
  }
  @media screen and (max-width: 450px) {
    .dashboard-content {
      margin: 1.4em 0em 3.5em;
    }
    .button-nav {
      margin: 2.5em 1em 0;
    }
  }
</style>
