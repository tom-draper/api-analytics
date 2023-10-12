<script lang="ts">
  import type { NotificationState } from '../../lib/notification';

  export let state: NotificationState;
</script>

<div
  class="notification"
  class:show={state.show}
  class:error={state.style === 'error'}
  class:warn={state.style === 'warn'}
  class:success={state.style === 'success'}
>
  <div class="message">{state.message}</div>
</div>

<style scoped>
  .notification {
    position: fixed;
    right: 2em;
    bottom: 2em;
    display: grid;
    place-items: center;
    padding: 1.6em 3em;
    border-radius: 6px;
    opacity: 0;
    transition: opacity 0.2s linear;
  }

  .show {
    opacity: 1 !important;
  }

  .error {
    background: #521e24c7;
    background: #4c1b1fe6;
    color: #cc3444;
  }
  .warn {
    background: rgba(203, 203, 108, 0.9);
    color: yellow;
  }
  .success {
    background: #2f423af2;
    color: var(--highlight);
  }
</style>
