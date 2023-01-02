/// <reference types="svelte" />

type RequestsData = {
    api_key: string,
    request_id: number,
    hostname: string,
    ip_address: string,
    path: string,
    response_time: number,
    status: number,
    user_agent: string,
    method: number,
    framework: number
    created_at: string
}[]