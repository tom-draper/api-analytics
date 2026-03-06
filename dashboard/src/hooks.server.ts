import { error } from '@sveltejs/kit';
import { dev } from '$app/environment';
import type { Handle } from '@sveltejs/kit';

const HOSTED_PATHS = [
	'/faq',
	'/frequently-asked-questions',
	'/outages',
	'/privacy-policy',
];

export const handle: Handle = async ({ event, resolve }) => {
	const isHostedPath = HOSTED_PATHS.some((p) => event.url.pathname.startsWith(p));
	if (isHostedPath) {
		const host = event.request.headers.get('host') ?? '';
		const isHosted = dev || host.includes('apianalytics.dev');
		if (!isHosted) error(404);
	}
	return resolve(event);
};
