export function cachedFunction<T, R>(func: (input: T) => R): (input: T) => R {
    const cache = new Map<T, R>();
    return (input: T) => {
        if (cache.has(input)) return cache.get(input)!;
        const result = func(input);
        cache.set(input, result);
        return result;
    };
}