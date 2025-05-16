import { ColumnIndex, methodMap } from '$lib/consts';

export type EndpointFilterType = 'all' | 'redirect' | 'success' | 'client' | 'server';

export interface Endpoint {
    path: string;
    status: number;
    count: number;
}

export interface EndpointsResult {
    endpoints: Endpoint[];
    maxCount: number;
}

/**
 * Creates a frequency map of endpoints from request data
 */
function createEndpointFrequencyMap(data: RequestsData, ignoreParams: boolean): Map<string, Endpoint> {
    const freq = new Map<string, Endpoint>();
    
    for (const row of data) {
        // Create groups of endpoints by path + status
        const path = ignoreParams ? row[ColumnIndex.Path].split('?')[0] : row[ColumnIndex.Path];
        const status = row[ColumnIndex.Status];
        const endpointID = `${path}${status}`;
        
        let endpoint = freq.get(endpointID);
        if (!endpoint) {
            const method = methodMap[row[ColumnIndex.Method]];
            endpoint = {
                path: `${method}  ${path}`,
                status: status,
                count: 0
            };
            freq.set(endpointID, endpoint);
        }
        endpoint.count++;
    }
    
    return freq;
}

/**
 * Determines if a status code matches the active filter
 */
function statusMatchesFilter(status: number, activeFilter: EndpointFilterType): boolean {
    return (
        activeFilter === 'all' ||
        (activeFilter === 'success' && status >= 200 && status <= 299) ||
        (activeFilter === 'redirect' && status >= 300 && status <= 399) ||
        (activeFilter === 'client' && status >= 400 && status <= 499) ||
        (activeFilter === 'server' && status >= 500)
    );
}

/**
 * Returns filtered and sorted endpoints and the maximum count
 */
export function getEndpoints(
    data: RequestsData, 
    activeFilter: EndpointFilterType, 
    ignoreParams: boolean
): EndpointsResult {
    const freq = createEndpointFrequencyMap(data, ignoreParams);

    // Convert to array and filter by status
    const endpoints: Endpoint[] = [];
    let maxCount = 0;
    
    for (const endpoint of freq.values()) {
        if (statusMatchesFilter(endpoint.status, activeFilter)) {
            endpoints.push(endpoint);
            if (endpoint.count > maxCount) {
                maxCount = endpoint.count;
            }
        }
    }

    // Sort by count descending
    endpoints.sort((a, b) => b.count - a.count);

    return {
        endpoints: endpoints.slice(0, 50), // Limit to top 50
        maxCount
    };
}