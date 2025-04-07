
export function statusSuccess(status: number) {
    return status >= 200 && status <= 299
}

export function statusRedirect(status: number) {
    return status >= 300 && status <= 399
}

export function statusBad(status: number) {
    return status >= 400 && status <= 499
}

export function statusError(status: number) {
    return status >= 500
}