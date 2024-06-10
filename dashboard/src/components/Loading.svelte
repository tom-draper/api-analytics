<div class="loading-container">
	<img class="main" src="../img/logo.png" alt="" />
	<img class="muted" src="../img/logo-grey.png" alt="" />
</div>

<style scoped>
	.loading-container {
		width: 25px;
		position: relative;
	}
	.main {
		position: absolute;
	}
	.muted {
		position: absolute;
		clip: rect(0, 500px, 0, 0);
		animation-name: rotate;
		animation-timing-function: ease-out;
		animation-duration: 2s;
		animation-iteration-count: 10;
	}
	img {
		width: 100%;
	}

	.muted:hover {
		clip: rect(0, 25px, 45px, 0);
	}

	@keyframes rotate {
		0% {
			clip: rect(0, 25px, 45px, 0);
		}
		100% {
			clip: rect(0, 55px, 0, 0);
		}
	}
</style>
