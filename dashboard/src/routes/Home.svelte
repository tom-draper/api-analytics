<div class="home">
  <div class="landing-page-container">
    <div class="landing-page">
      <div class="left">
        <h1>API Analytics</h1>
        <h2>Monitoring and analytics for API frameworks.</h2>
        <div class="links">
          <a href="/generate" class="link"> <div class="text">Try now - it's <span class="italic">free</span></div></a>
          <a href="/generate" class="link secondary"> <div class="text">Demo</div></a>
        </div>
        <!-- <div class="lower-links">
          <a href="/dashboard" class="link lower-link tertiary"> <div class="text">Dashboard</div></a>
          <a href="/monitoring" class="link lower-link tertiary"> <div class="text">Monitoring</div></a>
        </div> -->
      </div>
      <div class="right">
        <img class="logo" src="img/home-logo2.png" alt="" />
      </div>
    </div>
  </div>
  <div class="dashboard">
    <div class="dashboard-title-container">
      <img class="lightning-top" src="img/logo.png" alt="" />
      <h1 class="dashboard-title">Dashboard</h1>
    </div>
    <div class="dashboard-content">
      <div class="dashboard-content-text">
        An all-in-one analytics dashboard. Real-time insight into your API's usage.
      </div>
      <div class="dashboard-btn-container">
        <a href="/dashboard" class="dashboard-btn secondary">
          <div class="dashboard-btn-text">
            Open
          </div>
        </a>
      </div>
    </div>
    <img class="dashboard-img" src="img/dashboard.png" alt="" />
  </div>
  <div class="dashboard">
    <div class="dashboard-title-container">
      <img class="lightning-top" src="img/logo.png" alt="" />
      <h1 class="dashboard-title">Monitoring</h1>
    </div>
    <div class="dashboard-content">
      <div class="dashboard-content-text">
        Active monitoring and error notifications. Peace of mind.
      </div>
      <div class="dashboard-btn-container">
        <a href="/" class="dashboard-btn secondary">
          <div class="dashboard-btn-text">
            Coming Soon
          </div>
        </a>
      </div>
    </div>
    <img class="dashboard-img" src="img/monitoring.png" alt="" />
  </div>
</div>

<style scoped>
  .landing-page {
    display: flex;
    place-items: center;
    padding-bottom: 6em;
  }
  .landing-page-container {
    display: grid;
    margin: 0 12% 0 15%;
    min-height: 100vh;
  }
  .left {
    flex: 1;
    text-align: left;
  }
  .right {
    display: grid;
    place-items: center;
  }
  /* .left,
  .right {
    padding-bottom: 6em;
  } */
  .logo {
    max-width: 1400px;
    width: 700px;
    margin-bottom: -50px;
  }

  h1 {
    font-size: 3.4em;
    /* font-family: 'Zilla Slab' */
  }

  h2 {
    color: white;
  }

  .links {
    color: #707070;
    display: flex;
    /* width:  400px; */
    margin-top: 30px;
    text-align: left;
  }
  .lower-links {
    margin-top: 60px;
    display: flex;
    flex-direction: column;
  }
  .lower-link {
    margin-bottom: 20px;
    width: 200px;
  }
  .link {
    width: fit-content;
    /* margin: 10px auto; */
    margin-right: 20px;
    font-size: 0.9;
  }
  .link:hover {
    background: #31aa73;
  }

  .secondary {
    background: #1c1c1c;
    border: 3px solid var(--highlight);
    color: var(--highlight);
  }
  .secondary:hover {
    background: #222222;
    /* background: var(--highlight); */
    /* color: black; */
  }

  .tertiary {
    background: rgb(201, 201, 201);
    width: 250px;
    display: grid;
    place-items: center;
  }

  .lightning-top {
    height: 70px;
    /* margin-top: -30px; */
  }

  .dashboard-title-container {
    display: flex;
    margin: 2em 4em;
  }

  a {
    background: var(--highlight);
    color: black;
    padding: 10px 20px;
    border-radius: 4px;
  }
  .italic {
    font-style: italic;
  }
  .dashboard {
    border: 3px solid var(--highlight);
    width: 80%;
    border-radius: 10px;
    margin: auto;
    /* background: var(--highlight); */
    /* height: 100vh; */
    margin-bottom: 8em;
    overflow: hidden;
    position: relative;
  }
  .dashboard-title {
    font-size: 2.5em;
    text-align: left;
    margin: 0.2em 1em auto 1em;
  }
  .dashboard-title-container {
    
    place-content: center;
  }
  .dashboard-img {
    width: 81%;
    /* position: absolute; */
    /* right: 9%; */
    /* bottom: -1%; */
    border-radius: 10px;
    /* border: 2px solid var(--highlight); */
    box-shadow: 0px 24px 120px -25px var(--highlight);
    margin-bottom: -1%;
  }
  .dashboard-content {
    text-align: left;
    color: white;
    margin: 0 5em 4em;
  }
  .dashboard-content-text {
    font-size: 1.1em;
    text-align: center;
  }
  .dashboard-btn-container {
    display: flex;
    justify-content: center;
    margin-top: 2em
  }
  .dashboard-btn-text {
    /* padding: 12px 20px; */
    width: 100px;
    text-align: center;
  }
</style>
