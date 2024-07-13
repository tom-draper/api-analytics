
export function getSourceURL() {
    const url = new URL(window.location.href);
    const source = url.searchParams.get('source');
    if (source === '') {
        return null;
    }
    return source;
}