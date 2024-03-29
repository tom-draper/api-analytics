export type Period =
    | '24 hours'
    | 'Week'
    | 'Month'
    | '3 months'
    | '6 months'
    | 'Year'
    | 'All time'

export type DashboardSettings = {
    disable404: boolean
    hostname: string | null
    period: Period
    targetEndpoint: {
        path: string | null,
        status: number | null,
    }
    targetLocation: string | null,
    hiddenEndpoints: Set<string>
}

export function initSettings(): DashboardSettings {
    return {
        disable404: false,
        hostname: null,
        period: 'Month',
        targetEndpoint: {
            path: null, 
            status: null
        },
        targetLocation: null,
        hiddenEndpoints: new Set(),
    }
}
