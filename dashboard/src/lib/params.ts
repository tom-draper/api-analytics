import { replaceState } from '$app/navigation';
import { page } from '$app/state';

export function setParam(param: string, value: string | null) {
	if (value === null) {
		page.url.searchParams.delete(param);
	} else {
		page.url.searchParams.set(param, value);
	}
	replaceState(page.url, page.state);
}

export function setParamNoReplace(param: string, value: string | null) {
	if (value === null) {
		page.url.searchParams.delete(param);
	} else {
		page.url.searchParams.set(param, value);
	}
}

/**
 * Toggles a selection target and keeps the matching URL param in sync: clicking
 * the already-selected value clears it, otherwise it becomes the new selection.
 * `set` writes the bindable target ($state can't be assigned through a plain ref).
 */
export function toggleParam<T extends string | number>(
	key: string,
	value: T,
	current: T | null,
	set: (next: T | null) => void
) {
	if (current === value) {
		set(null);
		setParam(key, null);
	} else {
		set(value);
		setParam(key, String(value));
	}
}
