import assert from "node:assert/strict";
import test from "node:test";

import { Analytics } from "./index.js";

const request = {
	hostname: "example.test",
	ip_address: null,
	user_agent: "test",
	path: "/",
	status: 200,
	method: "GET",
	response_time: 1,
	user_id: null,
	created_at: "2026-01-01T00:00:00.000Z",
};

test("flush retains a batch when the server rejects it", async (t) => {
	const originalFetch = globalThis.fetch;
	const originalError = console.error;
	t.after(() => {
		globalThis.fetch = originalFetch;
		console.error = originalError;
	});
	globalThis.fetch = async () => new Response("unavailable", { status: 503 });
	console.error = () => {};

	const client = new Analytics("key", "test");
	client.requests.push(request);
	await client.flush();

	assert.deepEqual(client.requests, [request]);
});

test("flush retains a batch after a network failure", async (t) => {
	const originalFetch = globalThis.fetch;
	const originalError = console.error;
	t.after(() => {
		globalThis.fetch = originalFetch;
		console.error = originalError;
	});
	globalThis.fetch = async () => {
		throw new Error("network unavailable");
	};
	console.error = () => {};

	const client = new Analytics("key", "test");
	client.requests.push(request);
	await client.flush();

	assert.deepEqual(client.requests, [request]);
});
