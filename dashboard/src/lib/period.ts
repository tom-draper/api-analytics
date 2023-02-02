export default function periodToDays(period: string): number {
    if (period == "24-hours") {
        return 1;
    } else if (period == "week") {
        return 7;
    } else if (period == "month") {
        return 30;
    } else if (period == "3-months") {
        return 30 * 3;
    } else if (period == "6-months") {
        return 30 * 6;
    } else if (period == "year") {
        return 365;
    } else {
        return null;
    }
}