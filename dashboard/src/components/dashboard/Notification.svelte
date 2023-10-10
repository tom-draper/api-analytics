<script lang="ts">
    export let message: string, style: "error" | "warn" | "success", show: boolean;
</script>

<div class="notification"
    class:show={show}
    class:error={style==="error"}
    class:warn={style==="warn"}
    class:success={style==="success"}
    >
    <div class="message">{message}</div>
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
        transition: opacity 0.6s linear;
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