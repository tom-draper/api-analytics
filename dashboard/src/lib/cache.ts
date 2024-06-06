export function cachedFunction(func: (input: any) => any) {
    const cache = new Map();
  
    return (input: any) => {
        if (cache.has(input)) {
            return cache.get(input);
        }
        const result = func(input);
        cache.set(input, result);
        return result
    }
}