<script lang="ts">
	export let title: string, description: string, img: string, path: string;
</script>

<div class="dashboard">
	<div class="dashboard-title-container">
		<img class="lightning-top" src="img/logo.png" alt="" />
		<h1 class="dashboard-title">{title}</h1>
	</div>
	<div class="dashboard-content">
		<div class="dashboard-content-text">
			{description}
		</div>
		<div class="dashboard-btn-container">
			<a href={path} class="dashboard-btn secondary">
				<div class="dashboard-btn-text">Open</div>
			</a>
		</div>
	</div>
	<img class="dashboard-img" src="img/{img}" alt="" />
</div>

<style scoped>
	h1 {
		font-size: 3.4em;
	}

	.secondary {
		background: var(--background);
		border: 3px solid var(--highlight);
		color: var(--highlight);
	}
	.secondary:hover {
		background: var(--dark-green);
	}

	.lightning-top {
		height: 60px;
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

	.dashboard {
		border: 3px solid var(--highlight);
		width: 80%;
		border-radius: 10px;
		margin: auto;
		margin-bottom: 8em;
		overflow: hidden;
		position: relative;
		z-index: 2;
	}
	.dashboard-title {
		font-size: 2.5em;
		text-align: left;
		margin: 0.05em 0.6em auto 0.6em;
	}
	.dashboard-title-container {
		place-content: center;
	}
	.dashboard-img {
		width: 81%;
		border-radius: 10px;
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
		margin-top: 2em;
	}
	.dashboard-btn-text {
		text-align: center;
	}

	@media screen and (max-width: 1200px) {
		.dashboard {
			width: 90%;
			margin-bottom: 4em;
		}
	}

	@media screen and (max-width: 700px) {
		h1 {
			font-size: 2.5em;
		}
		.dashboard {
			font-size: 0.8em;
		}

		.lightning-top {
			height: 40px;
			margin-top: auto;
		}
		.dashboard-title {
			margin: 0.2em 0.5em auto 0.7em;
		}
		.dashboard-img {
			margin-bottom: -3%;
		}
		.dashboard-content {
			margin: 0 2em 4em;
		}
	}
</style>
