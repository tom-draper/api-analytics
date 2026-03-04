import { serverURL } from './consts';
import { page } from '$app/state'

function getSourceURL() {
	const params = page.url.searchParams;
	const source = params.get('source');
	if (source === '' || source === null) {
		return null;
	}
	return cleanURL(source);
}

function getEnvSourceURL() {
	const url = import.meta.env.VITE_SERVER_URL;
	if (!url) return null;
	return cleanURL(url);
}

function cleanURL(url: string) {
	if (url.endsWith('/')) {
		return url.slice(0, -1);
	}
	return url;
}

export function getServerURL() {
	return getSourceURL() ?? getEnvSourceURL() ?? serverURL;
}
