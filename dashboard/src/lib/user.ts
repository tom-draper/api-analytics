import { ColumnIndex } from './consts';

export function getUserIdentifier(request: RequestsData[number]) {
	return (
		request[ColumnIndex.IPAddress] ??
		'' + request[ColumnIndex.UserID].toString() ??
		''
	);
}

export function formatUserID(ipAddress: string, customUserID: string) {
	return `${ipAddress}||${customUserID}`;
}

export function userTargeted(targetUser: { ipAddress: string, userID: string, composite: boolean }, ipAddress: string, userID: string) {
	if (!targetUser) {
		return false;
	}

	if (targetUser.composite) {
		return (
			formatUserID(ipAddress, userID) === formatUserID(targetUser.ipAddress, targetUser.userID)
		);
	} else {
		return targetUser.userID === userID;
	}
}

export function formatDisplayUserID(
	targetUser: { ipAddress: string; userID: string; composite: boolean } | null
) {
	if (!targetUser) {
		return '';
	}

	if (targetUser.composite && targetUser.ipAddress && targetUser.userID) {
		return `${targetUser.ipAddress} + ${targetUser.userID}`;
	} else {
		return targetUser.userID || targetUser.ipAddress || '';
	}
}