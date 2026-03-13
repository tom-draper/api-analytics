<script lang="ts">
</script>

<div class="nginx-section">
	<div class="nginx-card">
		<div class="nginx-logo">
			<img src="images/logos/nginx-analytics.svg" alt="NGINX Analytics" class="logo-img" />
		</div>
		<div class="nginx-content">
			<h2>Using NGINX?</h2>
			<p class="description">
				Try <a href="https://github.com/tom-draper/nginx-analytics" target="_blank" class="highlight">NGINX Analytics</a>, a flexible and fully self-hosted
				solution for real-time monitoring.
			</p>
			<a href="https://github.com/tom-draper/nginx-analytics" target="_blank" class="get-started">Get Started</a>
		</div>
	</div>
</div>

<style scoped>
	.nginx-section {
		display: flex;
		justify-content: center;
		padding: 2em 10% 5em;
	}

	.nginx-card {
		display: flex;
		align-items: center;
		gap: 2em;
		background-color: var(--dark-background);
		border-radius: var(--radius-lg);
		padding: 2em 2.5em;
		max-width: 700px;
		width: 100%;
	}

	.nginx-logo {
		flex-shrink: 0;
	}

	.logo-img {
		height: 5em;
		width: auto;
	}

	.nginx-content {
		flex: 1;
		text-align: left;
	}

	h2 {
		font-size: 1.2em;
		font-weight: bold;
		color: white;
		margin-bottom: 0.5em;
	}

	.description {
		color: var(--faint-text);
		font-size: 0.9em;
		margin: 0 0 1.2em;
		max-width: none;
	}

	.highlight {
		color: var(--highlight);
		font-weight: 500;
	}

	.get-started {
		display: inline-block;
		background: var(--highlight);
		color: black;
		padding: 10px 20px;
		border-radius: var(--radius-md);
		font-weight: 600;
		font-size: 0.9em;
	}

	.get-started:hover {
		background: #31aa73;
	}

	@media screen and (max-width: 700px) {
		.nginx-card {
			flex-direction: column;
		}

		.nginx-content {
			text-align: center;
		}
	}
</style>
