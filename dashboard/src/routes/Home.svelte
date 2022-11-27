<div class="home">
<img class="logo" src="img/home-logo.png" alt="" />
<h1>API Analytics</h1>

<div class="links">
  <a href="/generate" class="link"> <div class="text">API key</div></a>
  <a href="/generate" class="link"> <div class="text">Dashboard</div></a>
</div>
</div>

<style scoped>
  .home {
    display: grid;
    place-items: center;
  }
  .logo {
    width: 700px;
    /* margin-top: 25vh; */
    /* margin-bottom: 10px */
    margin-bottom: -80px;
  }

  h1 {
    font-size: 2.5em;
  }

  .links {
    color: #707070;
    display: flex;
    width: 400px;
    margin: 30px auto auto;
    flex-direction: column;
  }
  .link {
    /* margin: 0 20px; */
    width: fit-content;
    margin: 10px auto;
    color: #707070;
    color: white;
    font-size: 0.9;
    /* width: 200px; */
  }
  .link:hover {
    color: var(--highlight);
  }
  /* .text {
    width: 200px;
  } */
</style>
