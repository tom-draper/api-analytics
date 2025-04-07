export function formatPath(path: string, query: string): string {
    return `${path}${query ? `?${query}` : ''}`;
}