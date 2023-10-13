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
    background: rgba(76, 27, 31, 0.95);
    color: #cc3444;
  }
  .warn {
    background: rgba(186, 186, 100, 0.95);
    color: yellow;
  }
  .success {
    background: rgba(47, 66, 58, 0.95);
    color: var(--highlight);
  }
</style>
