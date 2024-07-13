export type NotificationStyle = 'error' | 'warn' | 'success';

export type NotificationState = {
	message: string;
	style: NotificationStyle;
	show: boolean;
};
