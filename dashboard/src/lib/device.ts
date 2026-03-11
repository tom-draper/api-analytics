import type { Candidate } from './candidates';

export const clientCandidates: Candidate[] = [
	{ name: 'Curl', regex: /curl\//, matches: 0 },
	{ name: 'Postman', regex: /PostmanRuntime\//, matches: 0 },
	{ name: 'Insomnia', regex: /insomnia\//, matches: 0 },
	{ name: 'Python requests', regex: /python-requests\//, matches: 0 },
	{ name: 'Nodejs fetch', regex: /node-fetch\//, matches: 0 },
	{ name: 'Seamonkey', regex: /Seamonkey\//, matches: 0 },
	{ name: 'Firefox', regex: /Firefox\//, matches: 0 },
	{ name: 'Chrome', regex: /Chrome\//, matches: 0 },
	{ name: 'Chromium', regex: /Chromium\//, matches: 0 },
	{ name: 'aiohttp', regex: /aiohttp\//, matches: 0 },
	{ name: 'Python', regex: /Python\//, matches: 0 },
	{ name: 'Go http', regex: /[Gg]o-http-client\//, matches: 0 },
	{ name: 'Java', regex: /Java\//, matches: 0 },
	{ name: 'axios', regex: /axios\//, matches: 0 },
	{ name: 'Dart', regex: /Dart\//, matches: 0 },
	{ name: 'OkHttp', regex: /OkHttp\//, matches: 0 },
	{ name: 'Uptime Kuma', regex: /Uptime-Kuma\//, matches: 0 },
	{ name: 'undici', regex: /undici\//, matches: 0 },
	{ name: 'Lush', regex: /Lush\//, matches: 0 },
	{ name: 'Zabbix', regex: /Zabbix/, matches: 0 },
	{ name: 'Guzzle', regex: /GuzzleHttp\//, matches: 0 },
	{ name: 'Uptime', regex: /Better Uptime/, matches: 0 },
	{ name: 'GitHub Camo', regex: /github-camo/, matches: 0 },
	{ name: 'Ruby', regex: /Ruby/, matches: 0 },
	{ name: 'Node.js', regex: /node/, matches: 0 },
	{ name: 'Next.js', regex: /Next\.js/, matches: 0 },
	{ name: 'Vercel Edge Functions', regex: /Vercel Edge Functions/, matches: 0 },
	{ name: 'OpenAI Image Downloader', regex: /OpenAI Image Downloader/, matches: 0 },
	{ name: 'OpenAI', regex: /OpenAI/, matches: 0 },
	{ name: 'Tsunami Security Scanner', regex: /TsunamiSecurityScanner/, matches: 0 },
	{ name: 'iOS', regex: /iOS\//, matches: 0 },
	{ name: 'Safari', regex: /Safari\//, matches: 0 },
	{ name: 'Edge', regex: /Edg\//, matches: 0 },
	{ name: 'Opera', regex: /(OPR|Opera)\//, matches: 0 },
	{ name: 'Internet Explorer', regex: /(; MSIE |Trident\/)/, matches: 0 },
];

export const deviceCandidates: Candidate[] = [
	{ name: 'iPhone', regex: /iPhone/, matches: 0 },
	{ name: 'Android', regex: /Android/, matches: 0 },
	{ name: 'Samsung', regex: /Tizen\//, matches: 0 },
	{ name: 'Mac', regex: /Macintosh/, matches: 0 },
	{ name: 'Windows', regex: /Windows/, matches: 0 },
];

export const osCandidates: Candidate[] = [
	{ name: 'Windows 3.11', regex: /Win16/, matches: 0 },
	{ name: 'Windows 95', regex: /(Windows 95)|(Win95)|(Windows_95)/, matches: 0 },
	{ name: 'Windows 98', regex: /(Windows 98)|(Win98)/, matches: 0 },
	{ name: 'Windows 2000', regex: /(Windows NT 5.0)|(Windows 2000)/, matches: 0 },
	{ name: 'Windows XP', regex: /(Windows NT 5.1)|(Windows XP)/, matches: 0 },
	{ name: 'Windows Server 2003', regex: /(Windows NT 5.2)/, matches: 0 },
	{ name: 'Windows Vista', regex: /(Windows NT 6.0)/, matches: 0 },
	{ name: 'Windows 7', regex: /(Windows NT 6.1)/, matches: 0 },
	{ name: 'Windows 8', regex: /(Windows NT 6.2)/, matches: 0 },
	{ name: 'Windows 10/11', regex: /(Windows NT 10.0)/, matches: 0 },
	{ name: 'Windows NT 4.0', regex: /(Windows NT 4.0)|(WinNT4.0)|(WinNT)|(Windows NT)/, matches: 0 },
	{ name: 'Windows ME', regex: /Windows ME/, matches: 0 },
	{ name: 'OpenBSD', regex: /OpenBSD/, matches: 0 },
	{ name: 'SunOS', regex: /SunOS/, matches: 0 },
	{ name: 'Android', regex: /Android/, matches: 0 },
	{ name: 'Linux', regex: /(Linux)|(X11)/, matches: 0 },
	{ name: 'MacOS', regex: /(Mac_PowerPC)|(Macintosh)/, matches: 0 },
	{ name: 'QNX', regex: /QNX/, matches: 0 },
	{ name: 'iOS', regex: /iPhone OS/, matches: 0 },
	{ name: 'BeOS', regex: /BeOS/, matches: 0 },
	{ name: 'OS/2', regex: /OS\/2/, matches: 0 },
	{
		name: 'Search Bot',
		regex: /(APIs-Google)|(AdsBot)|(nuhk)|(Googlebot)|(Storebot)|(Google-Site-Verification)|(Mediapartners)|(Yammybot)|(Openbot)|(Slurp)|(MSNBot)|(Ask Jeeves\/Teoma)|(ia_archiver)/,
		matches: 0,
	},
];

/** Pure (non-mutating) label lookup — used by periodFilter to build UA ID sets. */
export function matchLabel(ua: string | null, candidates: Candidate[]): string {
	if (!ua) return 'Unknown';
	for (const c of candidates) {
		if (c.regex.test(ua)) return c.name;
	}
	return 'Other';
}
