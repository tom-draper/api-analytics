import { ColumnIndex } from './consts'

function convertLineToCSV(line: RequestsData[number]) {
    let str = '';
    for (let i = 0; i < line.length; i++) {
        if (i > 0) str += ',';
        if (i === ColumnIndex.UserAgent) {
            // Commas in user agent strings will break the CSV
            str += `"${line[i]}"`;
        } else {
            str += line[i];
        }
    }
    return str;
}

function convertToCSV(data: RequestsData, columns: string[]) {
    let str = columns.join(',') + '\r\n';
    for (let i = 0; i < data.length; i++) {
        const line = convertLineToCSV(data[i]);
        str += line + '\r\n';
    }
    return str;
};

export default function exportCSV(data: RequestsData, columns: string[]) {
    const csv = convertToCSV(data, columns);
    const exportedFilename = 'export.csv';
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