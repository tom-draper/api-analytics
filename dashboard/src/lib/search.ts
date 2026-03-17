import { ColumnIndex, methodMap } from './consts';

type Operator = 'eq' | 'gt' | 'lt' | 'gte' | 'lte' | 'glob' | 'contains';

type SearchTerm = {
	field: string;
	op: Operator;
	value: string | number;
	rawValue: string;     // original string after stripping operator prefix
	valueLower: string;
	numericValue: number;
	pattern?: RegExp;
};

function globToRegex(pattern: string): RegExp {
	const escaped = pattern
		.replace(/[.+^${}()|[\]\\]/g, '\\$&')
		.replace(/\*/g, '.*')
		.replace(/\?/g, '.');
	return new RegExp(`^${escaped}$`, 'i');
}

function parseValue(raw: string): Omit<SearchTerm, 'field'> {
	if (raw.startsWith('>=')) {
		const str = raw.slice(2);
		const n = parseFloat(str);
		return { op: 'gte', value: n, rawValue: str, valueLower: str.toLowerCase(), numericValue: n };
	}
	if (raw.startsWith('<=')) {
		const str = raw.slice(2);
		const n = parseFloat(str);
		return { op: 'lte', value: n, rawValue: str, valueLower: str.toLowerCase(), numericValue: n };
	}
	if (raw.startsWith('>')) {
		const str = raw.slice(1);
		const n = parseFloat(str);
		return { op: 'gt', value: n, rawValue: str, valueLower: str.toLowerCase(), numericValue: n };
	}
	if (raw.startsWith('<')) {
		const str = raw.slice(1);
		const n = parseFloat(str);
		return { op: 'lt', value: n, rawValue: str, valueLower: str.toLowerCase(), numericValue: n };
	}
	if (raw.includes('*') || raw.includes('?')) {
		return { op: 'glob', value: raw, rawValue: raw, valueLower: raw.toLowerCase(), numericValue: NaN, pattern: globToRegex(raw) };
	}
	return { op: 'eq', value: raw, rawValue: raw, valueLower: raw.toLowerCase(), numericValue: parseFloat(raw) };
}

export function parseSearch(query: string): SearchTerm[] {
	const terms: SearchTerm[] = [];
	const tokens = query.match(/(?:[^\s"]+|"[^"]*")+/g) ?? [];
	for (const token of tokens) {
		const colonIdx = token.indexOf(':');
		if (colonIdx === -1) {
			const v = token.toLowerCase();
			terms.push({ field: 'any', op: 'contains', value: v, rawValue: v, valueLower: v, numericValue: NaN });
			continue;
		}
		const field = token.slice(0, colonIdx).toLowerCase();
		const rawValue = token.slice(colonIdx + 1).replace(/^"|"$/g, '');
		terms.push({ field, ...parseValue(rawValue) });
	}
	return terms;
}

// Returns [from, to) in ms for a given date string.
// Supports: today, yesterday, Nd/Nw/Nm/Nh, 2024, 2024-01, 2024-01-15,
//           2024-01-15T14, 2024-01-15T14:30, 2024-01-15T14:30:00
function parseTimestampRange(raw: string): { from: number; to: number } | null {
	const s = raw.trim().toLowerCase();
	const now = new Date();
	const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());

	if (s === 'today') {
		return { from: today.getTime(), to: today.getTime() + 86400000 };
	}
	if (s === 'yesterday') {
		const y = today.getTime() - 86400000;
		return { from: y, to: today.getTime() };
	}
	// Relative: 7d, 2w, 3m, 6h
	const rel = s.match(/^(\d+)(h|d|w|m)$/);
	if (rel) {
		const n = parseInt(rel[1]);
		const unit = rel[2];
		const from = new Date(now);
		if (unit === 'h') from.setHours(from.getHours() - n);
		else if (unit === 'd') from.setDate(from.getDate() - n);
		else if (unit === 'w') from.setDate(from.getDate() - n * 7);
		else from.setMonth(from.getMonth() - n);
		return { from: from.getTime(), to: now.getTime() };
	}
	// Year: 2024
	if (/^\d{4}$/.test(s)) {
		const y = parseInt(s);
		return { from: new Date(y, 0, 1).getTime(), to: new Date(y + 1, 0, 1).getTime() };
	}
	// Month: 2024-01
	if (/^\d{4}-\d{2}$/.test(s)) {
		const [y, m] = s.split('-').map(Number);
		return { from: new Date(y, m - 1, 1).getTime(), to: new Date(y, m, 1).getTime() };
	}
	// Day: 2024-01-15
	if (/^\d{4}-\d{2}-\d{2}$/.test(s)) {
		const [y, m, d] = s.split('-').map(Number);
		return { from: new Date(y, m - 1, d).getTime(), to: new Date(y, m - 1, d + 1).getTime() };
	}
	// Hour: 2024-01-15T14 or 2024-01-15 14
	const hourMatch = raw.match(/^(\d{4}-\d{2}-\d{2})[T ](\d{2})$/i);
	if (hourMatch) {
		const [y, mo, d] = hourMatch[1].split('-').map(Number);
		const h = parseInt(hourMatch[2]);
		return { from: new Date(y, mo - 1, d, h).getTime(), to: new Date(y, mo - 1, d, h + 1).getTime() };
	}
	// Minute: 2024-01-15T14:30
	const minuteMatch = raw.match(/^(\d{4}-\d{2}-\d{2})[T ](\d{2}):(\d{2})$/i);
	if (minuteMatch) {
		const [y, mo, d] = minuteMatch[1].split('-').map(Number);
		const h = parseInt(minuteMatch[2]);
		const min = parseInt(minuteMatch[3]);
		return { from: new Date(y, mo - 1, d, h, min).getTime(), to: new Date(y, mo - 1, d, h, min + 1).getTime() };
	}
	// Second: 2024-01-15T14:30:00
	const secondMatch = raw.match(/^(\d{4}-\d{2}-\d{2})[T ](\d{2}):(\d{2}):(\d{2})$/i);
	if (secondMatch) {
		const [y, mo, d] = secondMatch[1].split('-').map(Number);
		const h = parseInt(secondMatch[2]);
		const min = parseInt(secondMatch[3]);
		const sec = parseInt(secondMatch[4]);
		return { from: new Date(y, mo - 1, d, h, min, sec).getTime(), to: new Date(y, mo - 1, d, h, min, sec + 1).getTime() };
	}
	return null;
}

function matchesTerm(
	row: RequestsData[number],
	term: SearchTerm,
	userAgents: Record<number, string>
): boolean {
	const { field, op, valueLower, numericValue, pattern, rawValue } = term;

	function stringMatch(cell: string): boolean {
		if (op === 'glob' && pattern) return pattern.test(cell);
		const lc = cell.toLowerCase();
		if (op === 'contains') return lc.includes(valueLower);
		return lc === valueLower;
	}

	function numericMatch(cell: number): boolean {
		if (op === 'gt') return cell > numericValue;
		if (op === 'gte') return cell >= numericValue;
		if (op === 'lt') return cell < numericValue;
		if (op === 'lte') return cell <= numericValue;
		return cell === numericValue;
	}

	switch (field) {
		case 'status': {
			const status = row[ColumnIndex.Status] as number;
			if (op === 'eq') {
				if (valueLower.endsWith('xx')) return String(status)[0] === valueLower[0];
				return status === numericValue;
			}
			return numericMatch(status);
		}
		case 'method': {
			const methodStr = methodMap[row[ColumnIndex.Method] as number] ?? '';
			return stringMatch(methodStr);
		}
		case 'path':
			return stringMatch((row[ColumnIndex.Path] as string) ?? '');
		case 'hostname':
			return stringMatch((row[ColumnIndex.Hostname] as string) ?? '');
		case 'location':
			return stringMatch((row[ColumnIndex.Location] as string) ?? '');
		case 'user_id':
		case 'userid':
			return stringMatch((row[ColumnIndex.UserID] as string) ?? '');
		case 'ip':
		case 'ip_address':
			return stringMatch((row[ColumnIndex.IPAddress] as string) ?? '');
		case 'response_time':
		case 'responsetime':
		case 'rt':
			return numericMatch(row[ColumnIndex.ResponseTime] as number);
		case 'user_agent':
		case 'useragent':
		case 'ua': {
			const uaStr = userAgents[row[ColumnIndex.UserAgent] as number] ?? '';
			return stringMatch(uaStr);
		}
		case 'referrer':
			return stringMatch((row[ColumnIndex.Referrer] as string) ?? '');
		case 'timestamp':
		case 'ts':
		case 'date':
		case 'time':
		case 'created_at': {
			const ts = (row[ColumnIndex.CreatedAt] as Date).getTime();
			const range = parseTimestampRange(rawValue);
			if (!range) return false;
			if (op === 'eq') return ts >= range.from && ts < range.to;
			// For operators, treat the range boundary semantically:
			// >date  → after the end of that period
			// >=date → from the start of that period
			// <date  → before the start of that period
			// <=date → up to (and including) the end of that period
			if (op === 'gt') return ts >= range.to;
			if (op === 'gte') return ts >= range.from;
			if (op === 'lt') return ts < range.from;
			if (op === 'lte') return ts < range.to;
			return false;
		}
		case 'any':
			return (
				stringMatch((row[ColumnIndex.Path] as string) ?? '') ||
				stringMatch((row[ColumnIndex.Hostname] as string) ?? '') ||
				stringMatch((row[ColumnIndex.IPAddress] as string) ?? '') ||
				stringMatch((row[ColumnIndex.UserID] as string) ?? '')
			);
		default:
			return true;
	}
}

export function applySearch(
	data: RequestsData,
	query: string,
	userAgents: Record<number, string>
): RequestsData {
	if (!query.trim()) return data;
	const terms = parseSearch(query);
	if (terms.length === 0) return data;
	return data.filter((row) => terms.every((term) => matchesTerm(row, term, userAgents)));
}
