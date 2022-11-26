<script lang="ts">
  import { onMount } from "svelte";

  function thisWeek(date: Date): boolean {
    let weekAgo = new Date();
    weekAgo.setDate(weekAgo.getDate() - 7);
    return date > weekAgo;
  }

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
      margin: { r: 35, l: 70, t: 10, b: 20, pad: 0 },
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
    "#E46161",  // Red
    "#EBEB81",  // Yellow
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

  // function build() {
  //   let totalRequests = 0;
  //   for (let i = 0; i < data.length; i++) {
  //     let date = new Date(data[i].created_at);
  //     if (thisWeek(date)) {
  //       totalRequests++;
  //     }
  //   }
  //   requestsPerHour = ((24 * 7) / totalRequests).toFixed(2);
  // }

  let plotDiv;
  onMount(() => {
    genPlot();
  });

  export let data: any;
</script>

<div class="card" title="Last week">
  <div class="card-title">OS</div>
  <div id="plotly">
    <div id="plotDiv" bind:this={plotDiv}>
      <!-- Plotly chart will be drawn inside this DIV -->
    </div>
  </div>
</div>

<style>
  .card {
    margin: 2em 2em 2em 0;
    padding-bottom: 1em;
  }
  #plotDiv {
    margin: 0 20px;
  }
</style>
