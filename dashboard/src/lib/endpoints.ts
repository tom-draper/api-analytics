export type EndpointFilterType = 'all' | 'redirect' | 'success' | 'client' | 'server';

export interface Endpoint {
	path: string;
	status: number;
	count: number;
}

/**
 * Determines if a status code matches the active filter
 */
export function statusMatchesFilter(status: number, activeFilter: EndpointFilterType): boolean {
	return (
		activeFilter === 'all' ||
		(activeFilter === 'success' && status >= 200 && status <= 299) ||
		(activeFilter === 'redirect' && status >= 300 && status <= 399) ||
		(activeFilter === 'client' && status >= 400 && status <= 499) ||
		(activeFilter === 'server' && status >= 500)
	);
}
