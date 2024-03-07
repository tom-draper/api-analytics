import { ColumnIndex } from './consts';

function convertLineToCSV(line: RequestsData[number], userAgents: UserAgents) {
    let str = '';
    for (let i = 0; i < line.length; i++) {
        if (i > 0) {
            str += ',';
        }
        if (i === ColumnIndex.UserAgent) {
            // Commas in user agent strings will break the CSV
            if (line[i] in userAgents) {
                str += `"${userAgents[line[i]]}"`;
            } else {
                str += `""`;
            }
        } else {
            str += line[i];
        }
    }
    return str;
}

function convertToCSV(
    data: RequestsData,
    columns: string[],
    userAgents: UserAgents
) {
    let str = columns.join(',') + '\r\n';
    for (let i = 0; i < data.length; i++) {
        const line = convertLineToCSV(data[i], userAgents);
        str += line + '\r\n';
    }
    return str;
}

export default function exportCSV(
    data: RequestsData,
    columns: string[],
    userAgents: UserAgents
) {
    const csv = convertToCSV(data, columns, userAgents);
    const exportedFilename = `api_analytics_${new Date()
        .toJSON()
        .replace(/[- .]/g, '_')}.csv`;
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });

    const link = document.createElement('a');
    if (link.download !== undefined) {
        // Browsers that support HTML5 download attribute
        const url = URL.createObjectURL(blob);
        link.setAttribute('href', url);
        link.setAttribute('download', exportedFilename);
        link.style.visibility = 'hidden';
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    }
}
