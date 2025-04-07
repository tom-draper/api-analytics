/// <reference lib="webworker" />
import generateDemoData from '$lib/demo';

self.onmessage = async () => {
	const data = generateDemoData();
	postMessage(data);
};