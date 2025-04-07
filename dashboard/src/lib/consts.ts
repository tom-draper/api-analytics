export const serverURL = 'https://www.apianalytics-server.com';

export const pageSize = 200_000;

export const columns = [
	'ip_address',
	'path',
	'hostname',
	'user_agent',
	'method',
	'response_time',
	'status',
	'location',
	'user_id',
	'time',
];

export const enum ColumnIndex {
	IPAddress = 0,
	Path = 1,
	Hostname = 2,
	UserAgent = 3,
	Method = 4,
	ResponseTime = 5,
	Status = 6,
	Location = 7,
	UserID = 8,
	CreatedAt = 9,
}

export const graphColors = [
	'#3FCF8E', // Green
	'#5784BA', // Blue
	'#EBEB81', // Yellow
	'#218B82', // Sea green
	'#FFD6A5', // Orange
	'#F9968B', // Salmon
	'#B1A2CA', // Purple
	'#E46161', // Red
];

// Integer to method string mapping used by server
export const methodMap = [
	'GET',
	'POST',
	'PUT',
	'PATCH',
	'DELETE',
	'OPTIONS',
	'CONNECT',
	'HEAD',
	'TRACE',
];