import { ColumnIndex, methodMap } from './consts';

type Operator = 'eq' | 'gt' | 'lt' | 'gte' | 'lte' | 'glob' | 'contains';

type SearchTerm = {
	field: string;
	op: Operator;
	value: string | number;
	valueLower: string;   // precomputed lowercase string value
	numericValue: number; // precomputed numeric value (NaN if not applicable)
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
		const n = parseFloat(raw.slice(2));
		return { op: 'gte', value: n, valueLower: String(n), numericValue: n };
	}
	if (raw.startsWith('<=')) {
		const n = parseFloat(raw.slice(2));
		return { op: 'lte', value: n, valueLower: String(n), numericValue: n };
	}
	if (raw.startsWith('>')) {
		const n = parseFloat(raw.slice(1));
		return { op: 'gt', value: n, valueLower: String(n), numericValue: n };
	}
	if (raw.startsWith('<')) {
		const n = parseFloat(raw.slice(1));
		return { op: 'lt', value: n, valueLower: String(n), numericValue: n };
	}
	if (raw.includes('*') || raw.includes('?')) {
		return { op: 'glob', value: raw, valueLower: raw.toLowerCase(), numericValue: NaN, pattern: globToRegex(raw) };
	}
	return { op: 'eq', value: raw, valueLower: raw.toLowerCase(), numericValue: parseFloat(raw) };
}

export function parseSearch(query: string): SearchTerm[] {
	const terms: SearchTerm[] = [];
	const tokens = query.match(/(?:[^\s"]+|"[^"]*")+/g) ?? [];
	for (const token of tokens) {
		const colonIdx = token.indexOf(':');
		if (colonIdx === -1) {
			terms.push({ field: 'any', op: 'contains', value: token.toLowerCase(), valueLower: token.toLowerCase(), numericValue: NaN });
			continue;
		}
		const field = token.slice(0, colonIdx).toLowerCase();
		const rawValue = token.slice(colonIdx + 1).replace(/^"|"$/g, '');
		terms.push({ field, ...parseValue(rawValue) });
	}
	return terms;
}

function matchesTerm(
	row: RequestsData[number],
	term: SearchTerm,
	userAgents: Record<number, string>
): boolean {
	const { field, op, valueLower, numericValue, pattern } = term;

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
