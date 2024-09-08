import { ColumnIndex } from './consts';

export function getUserIdentifier(request: RequestsData[number]) {
	return (
		request[ColumnIndex.IPAddress] ??
		'' + request[ColumnIndex.UserID].toString() ??
		''
	);
}
