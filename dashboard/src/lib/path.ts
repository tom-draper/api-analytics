export function formatPath(path: string, query: string): string {
    let basePath = import.meta.env.VITE_RELATIVE_DASHBOARD_URL || '';
    return `${basePath}${path}${query ? `?${query}` : ''}`;
}