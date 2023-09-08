/// <reference types="svelte" />

// ip_address, path, user_agent, method, response_time, status, location, created_at
type RequestsData = [string, string, string, number, number, number, string, string][]

type MonitorSample = {
    status: number;
    responseTime: number;
    createdAt: Date;
};

type MonitorData = { [url: string]: MonitorSample[] }
