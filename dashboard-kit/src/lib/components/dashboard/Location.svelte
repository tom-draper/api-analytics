<script lang="ts">
	import { replaceState } from '$app/navigation';
	import { page } from '$app/state';
	import { ColumnIndex } from '$lib/consts';

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

	function getLocationBars(data: RequestsData): LocationBar[] {
		let max = 0;
		const locationsFreq: ValueCount = {};
		for (let i = 0; i < data.length; i++) {
			const location = data[i][ColumnIndex.Location];
			if (!location) {
				continue;
			}
			if (location in locationsFreq) {
				locationsFreq[location]++;
			} else {
				locationsFreq[location] = 1;
			}
			if (locationsFreq[location] > max) {
				max = locationsFreq[location];
			}
		}

		const locationBars = Object.entries(locationsFreq)
			.map(([location, count]) => {
				return {
					location: location,
					frequency: count,
					height: count / max,
				};
			})
			.sort((a, b) => {
				return b.frequency - a.frequency;
			});

		return locationBars;
	}

	function setLocationParam(location: string | null) {
		if (location === null) {
			page.url.searchParams.delete('location')
		} else {
			page.url.searchParams.set('location', location);
		}
		replaceState(page.url, page.state);
	}

	let locations: LocationBar[] = [];

	$: if (data) {
		locations = getLocationBars(data);
	}

	export let data: RequestsData, targetLocation: string | null;
</script>

<div class="card">
	<div class="card-title">Location</div>
	{#if locations.length > 0}
		<div class="locations-count">{locations.length} location{locations.length > 1 ? 's' : ''}</div>
		<div class="bars">
			{#each locations.slice(0, 12) as location}
				<div class="bar-container">
					<button
						aria-label="location"
						class="bar"
						title="{countryCodeToName(
							location.location,
						)}: {location.frequency.toLocaleString()} requests"
						on:click={() => {
							const value = targetLocation === location.location ? null : location.location;
							targetLocation = value;
							setLocationParam(value);
						}}
					>
						<div
							class="bar-inner"
							style="height: {location.height * 100}%"
						></div>
					</button>
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
		height: 190px;
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
		height: 190px;
		display: grid;
		place-items: center;
		font-size: 0.95em;
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
