import { SERVER_URL } from './consts';
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
	try {
		return cleanURL(process.env.SERVER_URL);
	} catch (e) {
		return null;
	}
}

function cleanURL(url: string) {
	if (url.endsWith('/')) {
		return url.slice(0, -1);
	}
	return url;
}

export function getServerURL() {
	return getSourceURL() ?? getEnvSourceURL() ?? SERVER_URL;
}
