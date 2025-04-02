<script lang="ts">
	import Lightning from '$components/Lightning.svelte';
</script>

<div
	class="flex justify-center gap-24 pb-20 pt-4 max-lg:gap-16 max-md:gap-8 max-md:px-8 max-sm:flex-col max-sm:items-center max-sm:gap-8"
>
	<div class="price-card p-4">
		<div class="m-1 mb-4 w-6 text-[var(--light-background)]">
			<Lightning />
		</div>
		<h3>Hobby</h3>
		<div class="price mb-[2rem]">Free</div>
		<div class="mb-8"><a href="/generate">Get Started</a></div>
		<ul>
			<li class="feature text-[#ededed]">50,000 request storage</li>
			<li class="feature">Dashboard access</li>
			<li class="feature">Raw data access</li>
			<li class="feature">Up to 3 monitors</li>
		</ul>
	</div>
	<div class="price-card pro p-4">
		<div class="m-1 mb-4 w-6 text-[var(--highlight)]">
			<Lightning />
		</div>
		<h3>Pro</h3>
		<div class="price">$4.99/month</div>
		<div class=" mb-[1rem] text-xs">Cancel anytime</div>
		<div class="mb-8"><button>Upgrade</button></div>
		<ul>
			<li class="feature text-[#ededed]">2 million request storage</li>
			<li class="feature">Dashboard access</li>
			<li class="feature">Explorer access</li>
			<li class="feature">Raw data access</li>
			<li class="feature">Up to 20 monitors</li>
		</ul>
	</div>
</div>

<style scoped>
	.price-card {
		background-color: var(--dark-background);
		color: var(--faint-text);
		z-index: 50;
		border: 3px solid var(--dark-background);
		border-radius: 10px;
		text-align: left;
		padding: 1.4em 2em;
		width: 320px;
		padding-bottom: 10em;
		box-shadow: 0px 24px 120px -25px var(--highlight);
		box-shadow: 0px 20px 140px -25px var(--highlight);
	}
	.pro {
		border: 3px solid var(--highlight);
		border-radius: 10px;
	}
	h3 {
		color: #ededed;
		font-size: 1.5em;
	}
	.price {
		font-size: 1.2em;
		font-weight: 600;
		color: var(--highlight);
	}
	.feature {
		font-size: 0.85em;
		margin-bottom: 0.5em;
	}

	a,
	button {
		background-color: var(--highlight);
		color: var(--dark-background);
		border-radius: 4px;
		padding: 0.4em 1em;
		font-size: 0.9em;
		font-weight: 600;
		cursor: pointer;
	}
	.feature {
		font-size: 0.85em;
		margin-bottom: 0.5em;
	}

	li.feature {
		list-style: none;
		position: relative;
		padding-left: 1.2em;
	}

	li.feature::before {
		content: '';
		position: absolute;
		left: 0;
		top: 50%;
		transform: translateY(-50%);
		width: 6px;
		height: 6px;
		background-color: var(--highlight);
		border-radius: 50%;
	}
</style>
