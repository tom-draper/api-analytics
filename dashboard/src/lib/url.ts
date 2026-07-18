import { serverURL } from './consts';
import { page } from '$app/state';

function cleanURL(url: string) {
	return url.endsWith('/') ? url.slice(0, -1) : url;
}

/** Backend URL from the `?source=` query param (per-request override), if present. */
function getParamServerURL() {
	const source = page.url.searchParams.get('source');
	if (!source) return null;
	return cleanURL(source);
}

/** Backend URL baked in via the SERVER_URL env var when self-hosting the dashboard. */
function getEnvServerURL() {
	if (!__SERVER_URL__) return null;
	return cleanURL(__SERVER_URL__);
}

/**
 * Resolves the backend server URL, in order of precedence:
 *   1. the `?source=` URL param (lets anyone point the hosted dashboard at their own backend)
 *   2. the SERVER_URL env var (for a self-hosted dashboard)
 *   3. the default hosted backend
 */
export function getServerURL() {
	return getParamServerURL() ?? getEnvServerURL() ?? serverURL;
}

/** The backend this build trusts: the SERVER_URL env var, else the default hosted backend. The `?source=` override is deliberately excluded. */
function getTrustedServerURL() {
	return getEnvServerURL() ?? serverURL;
}

/**
 * Returns the origin the backend has been redirected to by `?source=` when it
 * differs from the trusted build-time backend, or null when the backend is
 * trusted. Pages that submit the secret API key use this to warn before sending
 * it to a non-default origin: a crafted `?source=` on the hosted dashboard would
 * otherwise exfiltrate a key the user types into a legitimate-looking page.
 */
export function untrustedBackendOrigin(): string | null {
	const source = getParamServerURL();
	if (!source) return null;

	let sourceOrigin: string;
	try {
		sourceOrigin = new URL(source).origin;
	} catch {
		return source; // unparseable source: treat as untrusted, show the raw value
	}

	try {
		if (sourceOrigin === new URL(getTrustedServerURL()).origin) return null;
	} catch {
		// trusted URL unparseable (should not happen) — fall through and warn.
	}
	return sourceOrigin;
}
