<div class="error">
	<div class="error-code">
		<div class="green-cover">
			<img src="img/logos/lightning-green.png" alt="" />
		</div>
		404
	</div>
	<div class="message">Page not found</div>
</div>

<style scoped>
	.error {
		margin-top: 8%;
		display: grid;
		place-items: center;
	}
	.error-code {
		font-size: 15em;
		font-weight: 900;
		color: var(--highlight);
		position: relative;
		width: fit-content;
	}
	.green-cover {
		background: var(--highlight);
		width: 0.3em;
		height: 0.6em;
		position: absolute;
		top: 30%;
		right: 40%;
		border-radius: 10px;
	}
	img {
		filter: brightness(0.12);
		width: 0.32em;
		position: absolute;
		right: 5%;
		top: 5%;
	}
	.message {
		color: var(--highlight);
		font-size: 1.4em;
	}
	@media screen and (max-width: 600px) {
		.error {
			font-size: 0.5em;
		}
	}
</style>
