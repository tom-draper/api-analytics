
export function toDay(date: Date) {
    return new Date(date.getFullYear(), date.getMonth(), date.getDate());
}

export function previousDay(date: Date) {
    return new Date(date.getFullYear(), date.getMonth(), date.getDate() - 1);
}

export function nextDay(date: Date) {
    return new Date(date.getFullYear(), date.getMonth(), date.getDate() + 1);
}

export function daysBetween(start: Date, end: Date) {
    const diffTime = Math.abs(end.getTime() - start.getTime());
    return Math.ceil(diffTime / (1000 * 60 * 60 * 24));
}
