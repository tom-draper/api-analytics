import { SERVER_URL } from './consts';

function getSourceURL() {
	const url = new URL(window.location.href);
	const source = url.searchParams.get('source');
	if (source === '') {
		return null;
	}
	return source;
}

function getEnvSourceURL() {
	try {
		return process.env.SERVER_URL;
	} catch (e) {
		return null;
	}
}

export function getServerURL() {
	return getSourceURL() ?? getEnvSourceURL() ?? SERVER_URL;
}
