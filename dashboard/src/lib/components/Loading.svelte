<script>
	import Lightning from "./Lightning.svelte";

</script>
<div class="lightning-wrapper">
	<!-- Background (darker version) -->
     <div class="lightning-bg text-[#1B5E4B]">
         <Lightning />
     </div>

    <!-- Foreground (lighter version) -->
     <div class="lightning-fill text-[var(--highlight)]">
         <Lightning />
     </div>
</div>

<style scoped>
	.lightning-wrapper {
		position: relative;
		height: auto;
	}

	.lightning-bg,
	.lightning-fill {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: auto;
	}

	.lightning-fill {
		animation: clipFill 1.5s ease-in-out infinite;
		clip-path: inset(100% 0 0 0);
	}

	@keyframes clipFill {
		0% {
			clip-path: inset(100% 0 0 0);
		}
		100% {
			clip-path: inset(0 0 0 0);
		}
	}
</style>
