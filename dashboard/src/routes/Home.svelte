<div class="home">
  <div class="left">
    <h1>API Analytics</h1>
    <h2>Monitoring and analytics for API frameworks.</h2>
    <div class="links">
      <a href="/generate" class="link"> <div class="text">Try now - it's <span class="italic">free</span></div></a>
      <a href="/dashboard" class="link secondary"> <div class="text">Dashboard</div></a>
      <a href="/monitoring" class="link secondary"> <div class="text">Monitoring</div></a>
    </div>
  </div>
  <div class="right">
    <img class="logo" src="img/home-logo2.png" alt="" />
  </div>
</div>

<style scoped>
  .home {
    display: flex;
    place-items: center;
    margin: 5em 12% 0 15%;
  }
  .left {
    flex: 1;
    text-align: left;
  }
  .right {
    display: grid;
    place-items: center;
  }
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
    background: var(--light-background);
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
</style>
