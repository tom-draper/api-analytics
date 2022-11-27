<div class="home">
<img class="logo" src="img/home-logo2.png" alt="" />
<h1>API Analytics</h1>

<div class="links">
  <a href="/generate" class="link"> <div class="text">API key</div></a>
  <a href="/dashboard" class="link"> <div class="text">Dashboard</div></a>
</div>
</div>

<style scoped>
  .home {
    display: grid;
    place-items: center;
  }
  .logo {
    width: 700px;
    margin-bottom: -50px;
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
    width: fit-content;
    margin: 10px auto;
    color: #707070;
    color: white;
    font-size: 0.9;
  }
  .link:hover {
    color: var(--highlight);
  }
</style>
