// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}

	type DashboardData = {
		userAgents: UserAgents;
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

	type RawMonitorSample = {
		status: number;
		response_time: number;
		created_at: Date | null;
	}

	type MonitorSample = {
		status: number;
		responseTime: number;
		createdAt: Date | null;
	};

	type MonitorData = { [url: string]: RawMonitorSample[] };

	type StatusLabel = "success" | "warning" | "error" | "no-request";

	// Monitor sample with label for status colour CSS class
	type Sample = MonitorSample & { label: StatusLabel };

	type ValueCount = {
		[value: string]: number;
	};
}

export { };
