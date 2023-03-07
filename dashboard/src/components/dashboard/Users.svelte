<script lang="ts">
  import { onMount } from "svelte";
  import periodToDays from "../../lib/period";

  function usersPlotLayout() {
    return {
      title: false,
      autosize: true,
      margin: { r: 0, l: 0, t: 0, b: 0, pad: 0 },
      hovermode: false,
      plot_bgcolor: "transparent",
      paper_bgcolor: "transparent",
      height: 60,
      yaxis: {
        gridcolor: "gray",
        showgrid: false,
        fixedrange: true,
        dragmode: false,
      },
      xaxis: {
        visible: false,
        dragmode: false,
      },
      dragmode: false,
    };
  }

  function lines() {
    let n = 5;
    let x = [...Array(n).keys()];
    let y = Array(n).fill(0);
    for (let i = 0; i < data.length; i++) {
      let idx = Math.floor(i / (data.length / n));
      if (data[i].ip_address != null && data[i].ip_address != "") {
        y[idx] += 1;
      }
    }
    return [
      {
        x: x,
        y: y,
        type: "lines",
        marker: { color: "transparent" },
        showlegend: false,
        line: { shape: "spline", smoothing: 1, color: "#3FCF8E30" },
        fill: "tozeroy",
        fillcolor: "#3fcf8e15",
      },
    ];
  }

  function usersPlotData() {
    return {
      data: lines(),
      layout: usersPlotLayout(),
      config: {
        responsive: true,
        showSendToCloud: false,
        displayModeBar: false,
      },
    };
  }

  function genPlot() {
    let plotData = usersPlotData();
    //@ts-ignore
    new Plotly.newPlot(
      plotDiv,
      plotData.data,
      plotData.layout,
      plotData.config
    );
  }

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
      if (data[i].ip_address != "" && data[i].ip_address != null) {
        users.add(data[i].ip_address);
      }
    }
    return users;
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
        usersPerHour = (numUsers / (24 * days)).toFixed(2);
      }
    } else {
      usersPerHour = "0";
    }
    genPlot();
  }

  let plotDiv: HTMLDivElement;
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

<button class="card" on:click={togglePeriod} title="Based on IP address">
  {#if perHour}
    <div class="card-title">
      Users <span class="per-hour">/ hour</span>
    </div>
    {#if usersPerHour != undefined}
      <div class="value">{usersPerHour}</div>
    {/if}
  {:else}
    {#if percentageChange != null}
      <div
        class="percentage-change"
        class:positive={percentageChange > 0}
        class:negative={percentageChange < 0}
      >
        {#if percentageChange > 0}
          <img class="arrow" src="../img/up.png" alt="" />
        {:else if percentageChange < 0}
          <img class="arrow" src="../img/down.png" alt="" />
        {/if}
        {Math.abs(percentageChange).toFixed(1)}%
      </div>
    {/if}
    <div class="card-title">Users</div>
    <div class="value">{numUsers.toLocaleString()}</div>
  {/if}
  <div id="plotly">
    <div id="plotDiv" bind:this={plotDiv}>
      <!-- Plotly chart will be drawn inside this DIV -->
    </div>
  </div>
</button>

<style scoped>
  .card {
    width: calc(215px - 1em);
    margin: 0 0 0 1em;
    position: relative;
    cursor: pointer;
    padding: 0;
    overflow: hidden;
  }
  .value {
    margin: 20px auto;
    width: fit-content;
    font-size: 1.8em;
    font-weight: 700;
    position: inherit;
    z-index: 2;
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
  .arrow {
    height: 11px;
  }
  #plotly {
    position: absolute;
    width: 110%;
    bottom: 0;
    overflow: hidden;
    margin: 0 -5%;
  }
  @media screen and (max-width: 1030px) {
    .card {
      width: auto;
      flex: 1;
      margin: 0 0 0 1em;
    }
  }
</style>
