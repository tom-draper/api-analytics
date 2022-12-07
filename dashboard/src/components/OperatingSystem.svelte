<script lang="ts">
  import { onMount } from "svelte";

  function getOS(userAgent: string): string {
    if (userAgent.match(/Win16/)) {
        return 'Windows 3.11'
    } else if (userAgent.match(/(Windows 95)|(Win95)|(Windows_95)/)) {
        return 'Windows 95'
    } else if (userAgent.match(/(Windows 98)|(Win98)/)) {
        return 'Windows 98'
    } else if (userAgent.match(/(Windows NT 5.0)|(Windows 2000)/)) {
        return 'Windows 2000'
    } else if (userAgent.match(/(Windows NT 5.1)|(Windows XP)/)) {
        return 'Windows XP'
    } else if (userAgent.match(/(Windows NT 5.2)/)) {
        return 'Windows Server 2003'
    } else if (userAgent.match(/(Windows NT 6.0)/)) {
        return 'Windows Vista'
    } else if (userAgent.match(/(Windows NT 6.1)/)) {
        return 'Windows 7'
    } else if (userAgent.match(/(Windows NT 6.2)/)) {
        return 'Windows 8'
    } else if (userAgent.match(/(Windows NT 10.0)/)) {
        return 'Windows 10'
    } else if (userAgent.match(/(Windows NT 4.0)|(WinNT4.0)|(WinNT)|(Windows NT)/)) {
        return 'Windows NT 4.0'
    } else if (userAgent.match(/Windows ME/)) {
        return 'Windows ME'
    } else if (userAgent.match(/OpenBSD/)) {
        return 'OpenBSE'
    } else if (userAgent.match(/SunOS/)) {
        return 'SunOS'
    } else if (userAgent.match(/(Linux)|(X11)/)) {
        return 'Linux'
    } else if (userAgent.match(/(Mac_PowerPC)|(Macintosh)/)) {
        return 'MacOS'
    } else if (userAgent.match(/QNX/)) {
        return 'QNX'
    } else if (userAgent.match(/BeOS/)) {
        return 'BeOS'
    } else if (userAgent.match(/OS\/2/)) {
        return 'OS/2'
    } else if (userAgent.match(/(nuhk)|(Googlebot)|(Yammybot)|(Openbot)|(Slurp)|(MSNBot)|(Ask Jeeves\/Teoma)|(ia_archiver)/)) {
        return 'Search Bot'
    } else {
        return 'Unknown'
    }
  }

  function osPlotLayout() {
    let monthAgo = new Date();
    monthAgo.setDate(monthAgo.getDate() - 30);
    let tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    return {
      title: false,
      autosize: true,
      margin: { r: 35, l: 70, t: 20, b: 20, pad: 0 },
      hovermode: "closest",
      plot_bgcolor: "transparent",
      paper_bgcolor: "transparent",
      height: 180,
      yaxis: {
        title: { text: "Requests" },
        gridcolor: "gray",
        showgrid: false,
        fixedrange: true,
      },
      xaxis: {
        visible: false,
      },
      dragmode: false,
    };
  }

  let colors = [
    "#3FCF8E",  // Green
    "#5784BA",  // Blue
    "#EBEB81",  // Yellow
    "#218B82",  // Sea green
    "#FFD6A5",  // Orange
    "#F9968B",  // Salmon
    "#B1A2CA",  // Purple
    "#E46161",  // Red
  ];

  function pieChart() {
    let osCount = {}
    for (let i = 0; i < data.length; i++) {
        let os = getOS(data[i].user_agent)
        if (!(os in osCount)) {
            osCount[os] = 0
        }
        osCount[os]++
    }

    let os = [];
    let count = [];
    for (let browser in osCount) {
        os.push(browser);
        count.push(osCount[browser])
    }
    return [{
    values: count,
        labels: os,
        type: 'pie',
          marker: {
            colors: colors
        },
    }];
  }

  function osPlotData() {
    return {
      data: pieChart(),
      layout: osPlotLayout(),
      config: {
        responsive: true,
        showSendToCloud: false,
        displayModeBar: false,
      },
    };
  }

  function genPlot() {
    let plotData = osPlotData();
    //@ts-ignore
    new Plotly.newPlot(
      plotDiv,
      plotData.data,
      plotData.layout,
      plotData.config
    );
  }

  let plotDiv: HTMLDivElement;
  onMount(() => {
    genPlot();
  });

  export let data: RequestsData;
</script>

<!-- <div class="card" title="Last week">
  <div class="card-title">OS</div> -->
  <div id="plotly">
    <div id="plotDiv" bind:this={plotDiv}>
      <!-- Plotly chart will be drawn inside this DIV -->
    </div>
  </div>
<!-- </div> -->

<style>
  .card {
    margin: 2em 2em 2em 0;
    flex: 1;
    padding-bottom: 1em;
  }
  #plotDiv {
    margin-right: 20px;
  }
  @media screen and (max-width: 1480px){
    .card {
        width: 100%;
      }
  }
</style>
