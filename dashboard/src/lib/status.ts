
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

export function statusUnknown(status: number) {
    return status < 200 || status > 599
}

export function statusSuccessful(status: number) {
    return statusSuccess(status) || statusRedirect(status)
}