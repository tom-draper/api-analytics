export type DashboardSettings = {
    disable404: boolean
}

export function initSettings(): DashboardSettings {
    return {
        disable404: false
    }
}