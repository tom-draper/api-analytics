/// <reference types="svelte" />

type DashboardData = {
	user_agents: UserAgents;
	requests: RequestsData;
};

// ip_address, path, hostname, user_agent, method, response_time, status, location, created_at
type RequestsData = [
	string,
	string,
	string,
	number,
	number,
	number,
	number,
	string,
	string,
	Date,
][];

type UserAgents = {
	[id: number]: string;
};

type MonitorSample = {
	status: number;
	responseTime: number;
	createdAt: Date;
};

// Monitor sample with label for status colour CSS class
type Sample = MonitorSample & { label: string };

type MonitorData = { [url: string]: MonitorSample[] };

type ValueCount = {
	[value: string]: number;
};