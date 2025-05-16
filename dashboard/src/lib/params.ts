import { replaceState } from "$app/navigation";
import { page } from "$app/state";

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