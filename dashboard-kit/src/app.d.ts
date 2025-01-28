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

	type MonitorData = { [url: string]: MonitorSample[] };

	type ValueCount = {
		[value: string]: number;
	};
}

export { };
