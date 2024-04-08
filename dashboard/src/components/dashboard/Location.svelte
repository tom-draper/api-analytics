<script lang="ts">
	import { onMount } from 'svelte';
	import { ColumnIndex } from '../../lib/consts';

	function getFlagEmoji(countryCode: string) {
		const codePoints = countryCode
			.toUpperCase()
			.split('')
			.map((char) => 127397 + char.charCodeAt(undefined));
		return String.fromCodePoint(...codePoints);
	}

	function countryCodeToName(countryCode: string) {
		const regionNames = new Intl.DisplayNames(['en'], { type: 'region' });
		return regionNames.of(countryCode);
	}

	type LocationBar = { location: string; frequency: number; height: number };

	function build() {
		let max = 0;
		const locationsFreq = {};
		for (let i = 0; i < data.length; i++) {
			const location = data[i][ColumnIndex.Location];
			if (!location) {
				continue;
			}
			locationsFreq[location] |= 0;
			locationsFreq[location] += 1;
			if (locationsFreq[location] > max) {
				max = locationsFreq[location];
			}
		}

		const locationBars = Object.keys(locationsFreq)
			.map((location) => {
				return {
					location: location,
					frequency: locationsFreq[location],
					height: locationsFreq[location] / max,
				};
			})
			.sort((a, b) => {
				return b.frequency - a.frequency;
			});

		locations = locationBars;
	}

	let locations: LocationBar[] = [];
	let mounted = false;
	onMount(() => {
		mounted = true;
	});

	$: data && mounted && build();

	export let data: RequestsData, targetLocation: string;
</script>

<div class="card">
	<div class="card-title">Location</div>
	{#if locations.length > 0}
		<div class="locations-count">{locations.length} locations</div>
		<div class="bars">
			{#each locations.slice(0, 12) as location}
				<div class="bar-container">
					<!-- svelte-ignore a11y-click-events-have-key-events -->
					<div
						class="bar"
						title="{countryCodeToName(
							location.location,
						)}: {location.frequency.toLocaleString()} requests"
						on:click={() => {
							if (targetLocation === location.location) {
								targetLocation = null;
							} else {
								targetLocation = location.location;
							}
						}}
					>
						<div
							class="bar-inner"
							style="height: {location.height * 100}%"
						/>
					</div>
					<div class="label">{getFlagEmoji(location.location)}</div>
				</div>
			{/each}
		</div>
	{:else}
		<div class="no-locations">
			<div class="no-locations-text">No Locations Found</div>
		</div>
	{/if}
</div>

<style scoped>
	.card {
		flex: 1.2;
		margin: 2em 1em 2em 0;
		position: relative;
	}
	.bars {
		height: 150px;
		display: flex;
		padding: 1.5em 2em 1em 2em;
	}

	.bar-container {
		flex: 1;
		margin: 0 5px;
		display: flex;
		flex-direction: column;
	}

	.no-locations {
		height: 180px;
		display: grid;
		place-items: center;
	}
	.no-locations-text {
		margin-bottom: 25px;
		color: #707070;
	}

	.bar {
		position: relative;
		flex-grow: 1;
		cursor: pointer;
		border-radius: 3px 3px 0 0;
	}
	.bar:hover {
		background: linear-gradient(transparent, #444);
	}

	.bar-inner {
		position: absolute;
		bottom: 0;
		width: 100%;
		background: var(--highlight);
		border-radius: 3px;
	}
	.label {
		padding-top: 8px;
	}
	.locations-count {
		position: absolute;
		top: 1.5em;
		right: 2em;
		font-size: 0.9em;
		color: #505050;
	}

	@media screen and (max-width: 1600px) {
		.card {
			width: 100%;
			margin: 2em 0 2em;
		}
	}
</style>
