<script lang="ts">
	import Prism from 'prismjs';

	import 'prismjs/components/prism-python.js';
	import 'prismjs/components/prism-typescript.js';
	import 'prismjs/components/prism-rust.js';
	import 'prismjs/components/prism-csharp.js';
	import 'prismjs/components/prism-go.js';
	import 'prismjs/components/prism-ruby.js';

	let { language, code }: { language: string; code: string } = $props();

	const highlightedCode = $derived(Prism.highlight(code, Prism.languages[language], language));
</script>

<svelte:head>
	<link rel="stylesheet" href="/css/prism.css" />
</svelte:head>

<pre class="language-{language} my-1 rounded !bg-[var(--dark-background)] !px-8 !py-[-4]">
    <code class="language-{language} block !text-sm font-normal">{@html highlightedCode}</code>
</pre>

<style>
	pre[class*='language-'] {
		padding: 0;
	}
</style>
