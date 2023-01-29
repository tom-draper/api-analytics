<div class="card">
  <img src="../img/logo.png" alt="" />
</div>

<style scoped>
  .card {
    width: calc(200px - 1em);
    margin: 0 1em 0 0;
    background: transparent;
    display: grid;
    place-items: center;
  }
  img {
    width: 25px;
  }
  @media screen and (max-width: 940px) {
    .card {
      width: auto;
      flex: 1;
      margin: 0 1em 0 0;
    }
  }
</style>
