<script lang="ts">
	import { onMount } from 'svelte';
	import animate from '$lib/animate';
	import { formatPath } from '$lib/path';

	let queryString: string = '';

	onMount(() => {
		// Start animation after delay
		animate();
		setInterval(animate, 8000);

		// Retrieve current URL parameters
		const params = new URLSearchParams(window.location.search);
		queryString = params.toString();
	});
</script>

<div class="landing-page-container">
	<div class="landing-page">
		<div class="text-left flex-1">
			<h1 class="font-bold my-8">API Analytics</h1>
			<h2 class="font-bold">Monitoring and analytics for API frameworks.</h2>
			<div class="links">
				<a
					href={formatPath('/generate', queryString)}
					class="link"
				>
					<div class="text">
						Try now – it's <span class="italic">free</span>
					</div>
				</a>
				<a
					href={formatPath('/dashboard/demo', queryString)}
					class="link secondary"
				>
					<div class="text">Demo</div>
				</a>
			</div>
		</div>
		<div class="right relative grid place-items-center">
			<img class="logo max-w-[1400px] w-[700px] mb-[-50px]" src="images/logos/home-logo.png" alt="" />
			<img
				id="hover-1"
				style="position: absolute;"
				class="logo animated"
				src="images/logos/floating-points-1.png"
				alt=""
			/>
			<img
				id="hover-2"
				style="position: absolute;"
				class="logo animated"
				src="images/logos/floating-points-2.png"
				alt=""
			/>
		</div>
	</div>
</div>

<style scoped>
	.landing-page {
		display: flex;
		place-items: center;
		padding-bottom: 6em;
	}
	.landing-page-container {
		display: grid;
		margin: 0 12% 0 15%;
		min-height: 100vh;
	}

	h1 {
		font-size: 3.4em;
	}

	h2 {
		color: white;
		font-size: 1.5em;
	}

	.links {
		color: var(--dim-text);
		display: flex;
		margin-top: 30px;
		text-align: left;
	}
	.link {
		width: fit-content;
		margin-right: 20px;
		font-size: 0.9;
	}
	.link:hover {
		background: #31aa73;
	}

	.secondary {
		background: var(--background);
		border: 3px solid var(--highlight);
		color: var(--highlight);
	}
	.secondary:hover {
		background: var(--dark-green);
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

	#hover-1 {
		transform: translateY(3.7%);
	}
	#hover-2 {
		transform: translateY(-3.7%);
	}

	.animated {
		transition: all 10s ease-in-out;
	}

	@media screen and (max-width: 1500px) {
		.landing-page-container {
			margin: 0 6% 0 7%;
		}
	}

	@media screen and (max-width: 1300px) {
		.landing-page-container {
			margin: 0 5% 0 6%;
		}
	}

	@media screen and (max-width: 1200px) {
		.landing-page {
			flex-direction: column-reverse;
		}
		.landing-page-container {
			margin: 0 2em;
		}
		.logo {
			width: 90%;
		}
	}

	@media screen and (max-width: 800px) {
		.right {
			margin-top: 2em;
		}
	}
	@media screen and (max-width: 700px) {
		h1 {
			font-size: 2.5em;
		}
		h2 {
			font-size: 1.2em;
		}
		.landing-page-container {
			min-height: unset;
			font-size: 0.8em;
		}
		.landing-page {
			padding-bottom: 8em;
		}
		.logo {
			margin-bottom: 0;
		}
	}

	@media screen and (max-width: 600px) {
		.logo {
			width: 100%;
		}
	}
</style>
