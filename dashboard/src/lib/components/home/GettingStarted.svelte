<script lang="ts">
	import frameworkExamples from '$lib/framework';
	import CodeHighlighter from '$components/home/CodeHighlighter.svelte';

	type Language = 'python' | 'javascript' | 'go' | 'rust' | 'ruby' | 'csharp' | 'php';
	type Framework =
		| 'FastAPI'
		| 'Flask'
		| 'Django'
		| 'Tornado'
		| 'Express'
		| 'Fastify'
		| 'Koa'
		| 'Gin'
		| 'Echo'
		| 'Fiber'
		| 'Chi'
		| 'Actix'
		| 'Axum'
		| 'Rocket'
		| 'Rails'
		| 'Sinatra'
		| 'ASP.NET Core'
		| 'Laravel';

	type SupportedFramework = {
		language: Language;
		framework: Framework;
	};

	const frameworks: SupportedFramework[] = [
		{ language: 'python', framework: 'FastAPI' },
		{ language: 'python', framework: 'Flask' },
		{ language: 'python', framework: 'Django' },
		{ language: 'python', framework: 'Tornado' },
		{ language: 'javascript', framework: 'Express' },
		{ language: 'javascript', framework: 'Fastify' },
		{ language: 'javascript', framework: 'Koa' },
		{ language: 'go', framework: 'Gin' },
		{ language: 'go', framework: 'Echo' },
		{ language: 'go', framework: 'Fiber' },
		{ language: 'go', framework: 'Chi' },
		{ language: 'rust', framework: 'Actix' },
		{ language: 'rust', framework: 'Axum' },
		{ language: 'rust', framework: 'Rocket' },
		{ language: 'ruby', framework: 'Rails' },
		{ language: 'ruby', framework: 'Sinatra' },
		{ language: 'csharp', framework: 'ASP.NET Core' }
	];
	let currentFramework = frameworks[0];

	function setFramework(value: SupportedFramework) {
		currentFramework = value;
	}
</script>

<div class="gradient-container">
	<div class="left-gradient gradient"></div>
	<div class="right-gradient gradient"></div>
	<div class="radial-dimmer dimmer"></div>
	<div class="linear-dimmer dimmer"></div>
</div>
<div class="add-middleware">
	<div class="add-middleware-title">Getting Started</div>
	<div class="frameworks">
		{#each frameworks as { language, framework }}
			<button
				class="framework {language}"
				class:active={currentFramework.framework === framework}
				on:click={() => {
					setFramework({ language, framework });
				}}>{framework}</button
			>
		{/each}
	</div>
	<div class="add-middleware-content">
		<div class="instructions-container">
			<div class="instructions">
				<div class="subtitle">Install</div>
				{#each frameworks as { framework }}
					<div class:hidden={currentFramework.framework !== framework}>
						<CodeHighlighter language="text" code={frameworkExamples[framework]?.install} />
					</div>
				{/each}
				<div class="subtitle">Add middleware to API</div>
				<div class="code-file">
					{frameworkExamples[currentFramework.framework].codeFile ?? ''}
				</div>
				{#each frameworks as { language, framework }}
					<div class:hidden={currentFramework.framework !== framework}>
						<CodeHighlighter {language} code={frameworkExamples[framework].example} />
					</div>
				{/each}
			</div>
		</div>
	</div>
</div>

<style scoped>
	.hidden {
		display: none;
	}
	.gradient-container {
		position: relative;
		display: flex;
		height: 800px;
		z-index: 1;
		margin-top: -15em;
	}
	.gradient {
		width: 100%;
		height: 100%;
		background: conic-gradient(from 270deg, var(--highlight), var(--dark-background));
		filter: brightness(1.05);
	}
	.left-gradient {
		margin-left: auto;
		transform: scale(-1, 1);
		margin-right: -1px;
	}
	.right-gradient {
		margin-right: auto;
		margin-left: -1px;
	}
	.dimmer {
		position: absolute;
		width: 100%;
		height: 100%;
	}
	.radial-dimmer {
		background: radial-gradient(transparent, var(--background));
	}
	.linear-dimmer {
		background: linear-gradient(var(--background), transparent, var(--background));
	}

	.display-framework {
		display: unset !important;
	}

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
	.left {
		flex: 1;
		text-align: left;
	}
	.right {
		display: grid;
		place-items: center;
	}
	.logo {
		max-width: 1400px;
		width: 700px;
		margin-bottom: -50px;
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

	.lightning-top {
		height: 65px;
	}

	.dashboard-title-container {
		display: flex;
		margin: 2em 4em;
	}

	.italic {
		font-style: italic;
	}
	.dashboard {
		border: 3px solid var(--highlight);
		width: 80%;
		border-radius: 10px;
		margin: auto;
		margin-bottom: 8em;
		overflow: hidden;
		position: relative;
	}
	.dashboard-title {
		font-size: 2.5em;
		text-align: left;
		margin: 0.2em 1em auto 1em;
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

	.add-middleware-title {
		color: var(--highlight);
		color: var(--background);
		font-size: 2.5em;
		font-weight: 800;
		margin-bottom: 1.5em;
	}
	.add-middleware {
		margin: auto;
		padding-bottom: 2em;
		border-radius: 6px;
		margin-top: -480px;
		z-index: 2;
		position: relative;
	}

	.frameworks {
		margin: 0 25%;
		overflow-x: auto;
	}
	.framework {
		color: var(--faint-text);
		/* color: black; */
		background: transparent;
		border: none;
		font-size: 1em;
		cursor: pointer;
		padding: 8px 13px;
		height: auto;
		width: auto;
		border: 3px solid transparent;
		border-radius: 4px;
	}
	.framework:hover {
		color: white;
	}
	.active {
		color: white;
	}
	.active.python {
		border: 3px solid #4b8bbe;
	}
	.active.go {
		border: 3px solid #00a7d0;
	}
	.active.javascript {
		border: 3px solid #edd718;
	}
	.active.rust {
		border: 3px solid #ef4900;
	}
	.active.ruby {
		border: 3px solid #cd0000;
	}
	.active.php {
		border: 3px solid #7377ad;
	}
	.active.csharp {
		border: 3px solid #178600;
	}
	.subtitle {
		color: var(--faint-text);
		margin: 10px 0 2px 16px;
		font-size: 0.85em;
	}
	.instructions-container {
		padding: 1.5em 2em 2em;
		width: 850px;
		margin: auto;
	}
	.instructions {
		text-align: left;
		display: flex;
		flex-direction: column;
		position: relative;
	}
	.code {
		display: none;
	}
	.code-file {
		position: absolute;
		font-size: 0.8em;
		top: 160px;
		color: rgb(97, 97, 97);
		text-align: right;
		right: 2.5em;
		margin-bottom: -2em;
	}

	.animated {
		transition: all 10s ease-in-out;
	}

	@media screen and (min-width: 1700px) {
		.frameworks {
			margin: 0 10%;
		}
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
		.dashboard {
			width: 90%;
			margin-bottom: 4em;
		}
		.logo {
			width: 100%;
		}
		.instructions-container {
			width: auto;
		}
	}

	@media screen and (max-width: 900px) {
		.home {
			font-size: 0.85em;
		}
		.code-file {
			top: 140px;
		}
		.add-middleware {
			margin-top: -465px;
		}
		.frameworks {
			margin: 0 5%;
		}
		.gradient-container {
			margin-top: -18em;
		}
	}
	@media screen and (max-width: 800px) {
		.home {
			font-size: 0.8em;
		}
		.right {
			margin-top: 2em;
		}
		.gradient-container {
			margin-top: -22em;
		}
	}
	@media screen and (max-width: 700px) {
		.landing-page-container {
			min-height: unset;
		}
		.landing-page {
			padding-bottom: 8em;
		}
		.lightning-top {
			height: 40px;
			margin-top: auto;
		}
		.dashboard-title {
			margin: 0.2em 0.5em auto 1em;
		}
		.dashboard-img {
			margin-bottom: -3%;
		}
		.dashboard-content {
			margin: 0 2em 4em;
		}
		.add-middleware-title {
			margin-bottom: 1em;
		}
		.instructions-container {
			padding: 0 4%;
		}
		.logo {
			margin-bottom: 0;
		}
		.gradient-container {
			margin-top: -18em;
			overflow: hidden;
		}

		.add-middleware {
			margin-top: -460px;
			font-size: 0.8em;
		}
		.left-gradient {
			width: 90%;
			margin-left: -40%;
		}
		.right-gradient {
			width: 90%;
			margin-right: -40%;
		}
	}

	@media screen and (max-width: 500px) {
		.gradient-container {
			margin-top: -18em;
			overflow: hidden;
		}
	}
</style>
