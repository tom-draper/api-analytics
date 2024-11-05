import { SERVER_URL } from './consts';

function getSourceURL() {
	const url = new URL(window.location.href);
	const source = url.searchParams.get('source');
	if (source === '') {
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
