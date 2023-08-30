
function getDemoStatus(date: Date, status: number): number {
    // Add down period
    let start = new Date();
    start.setDate(start.getDate() - 100);
    let end = new Date();
    end.setDate(end.getDate() - 96);

    if (date > start && date < end) {
        return 400;
    } else {
        return status;
    }
}

function getDemoResponseTime(date: Date, responseTime: number): number {
    // Add down period
    let start = new Date();
    start.setDate(start.getDate() - 100);
    let end = new Date();
    end.setDate(end.getDate() - 96);

    if (date > start && date < end) {
        return Math.floor(Math.random() * 400 + 200)
    } else {
        return responseTime
    }
}

function getDemoUserAgent(): string {
    let p = Math.random();
    if (p < 0.19) {
        return "Mozilla/5.0 (iPhone; CPU iPhone OS 13_5_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.1.1 Mobile/15E148 Safari/604.1";
    } else if (p < 0.3) {
        return "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0) Gecko/20100101 Firefox/47.0";
    } else if (p < 0.34) {
        return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59";
    } else if (p < 0.36) {
        return "curl/7.64.1";
    } else if (p < 0.39) {
        return "PostmanRuntime/7.26.5";
    } else if (p < 0.4) {
        return "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.106 Safari/537.36 OPR/38.0.2220.41";
    } else if (p < 0.42) {
        return "Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:42.0) Gecko/20100101 Firefox/42.0";
    } else if (p < 0.58) {
        return "python-requests/2.26.0";
    } else {
        return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36";
    }
}

function randomChoice(p: any[]): number {
    let rnd = p.reduce((a, b) => a + b) * Math.random();
    return p.findIndex((a) => (rnd -= a) < 0);
}

function randomChoices(p: number[], count: number): number[] {
    return Array.from(Array(count), randomChoice.bind(null, p));
}

function getHour(): number {
    let p = Array(24).fill(1);
    for (let i = 0; i < 3; i++) {
        for (let j = 5 + i; j < 11 - i; j++) {
            p[j] += 0.15;
        }
    }
    for (let i = 0; i < 4; i++) {
        for (let j = 11 + i; j < 21 - i; j++) {
            p[j] += 0.15;
        }
    }
    return randomChoices(p, 1)[0];
}

function getLocation(): string {
    const locations = ["GB", "US", "MX", "CA", "FR", "AU", "ID", "IE", "DE", "PL", "AF", "AL", "DZ", "AS", "AD", "AO", "AI", "AQ", "AG", "AR", "AM", "AW", "AT", "AZ", "BS", "BH", "BD", "BB", "BY", "BE", "BZ", "BJ", "BM", "BT", "BO", "BQ"]
    const p = [0.56, 1, 0.18, 0.2, 0.4, 0.3, 0.1, 0.05, 0.2, 0.06, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01]
    const idx = randomChoice(p)
    return locations[idx] as string
}

function addDemoSamples(
    demoData: RequestsData,
    endpoint: string,
    status: number,
    count: number,
    daysAgo: [number, number],
    responseTime: [number, number],
    user: [number, number],
) {
    for (let i = 0; i < count; i++) {
        let date = new Date();
        date.setDate(
            date.getDate() - Math.floor(Math.random() * daysAgo[1] + daysAgo[0])
        );
        date.setHours(getHour(), Math.floor(Math.random() * 60));
        demoData.push([
            Math.floor(Math.random() * user[1] + user[0]).toString(),
            endpoint,
            getDemoUserAgent(),
            0,
            getDemoResponseTime(date, Math.floor(Math.random() * responseTime[1] + responseTime[0])),
            getDemoStatus(date, status),
            getLocation(),
            date.toISOString()
        ]);
    }
}

export default function genDemoData(): RequestsData {
    let demoData: RequestsData = [];

    addDemoSamples(demoData, "/v1/", 200, 18000, [0, 650], [55, 240], [0, 3956]);
    addDemoSamples(demoData, "/v1/", 400, 1000, [0, 650], [55, 240], [0, 2041]);
    addDemoSamples(demoData, "/v1/account", 200, 8000, [0, 650], [55, 240], [0, 1854]);
    addDemoSamples(demoData, "/v1/account", 400, 1200, [0, 650], [55, 240], [0, 1654]);
    addDemoSamples(demoData, "/v1/help", 200, 700, [0, 650], [55, 240], [0, 1654]);
    addDemoSamples(demoData, "/v1/help", 400, 70, [0, 650], [55, 240], [0, 1654]);
    addDemoSamples(demoData, "/v2/", 200, 35000, [0, 650], [55, 240], [0, 1654]);
    addDemoSamples(demoData, "/v2/", 400, 200, [0, 650], [55, 240], [0, 1654]);
    addDemoSamples(demoData, "/v2/account", 200, 14000, [0, 650], [55, 240], [0, 1654]);
    addDemoSamples(demoData, "/v2/account", 400, 3000, [0, 650], [55, 240], [0, 1654]);
    addDemoSamples(demoData, "/v2/account/update", 200, 6000, [0, 650], [55, 240], [0, 1654]);
    addDemoSamples(demoData, "/v2/account/update", 400, 400, [0, 650], [55, 240], [0, 1654]);
    addDemoSamples(demoData, "/v2/help", 200, 6000, [0, 650], [55, 240], [0, 1654]);
    addDemoSamples(demoData, "/v2/help", 400, 400, [0, 650], [55, 240], [0, 1654]);

    addDemoSamples(demoData, "/v2/account", 200, 16000, [0, 450], [30, 100], [0, 1654]);
    addDemoSamples(demoData, "/v2/account", 400, 2000, [0, 450], [300, 1000], [0, 1654]);

    addDemoSamples(demoData, "/v2/help", 200, 8000, [0, 300], [30, 100], [0, 1654]);
    addDemoSamples(demoData, "/v2/help", 400, 800, [0, 300], [30, 100], [0, 1654]);

    addDemoSamples(demoData, "/v2/", 200, 4000, [0, 200], [30, 100], [0, 1654]);
    addDemoSamples(demoData, "/v2/", 400, 800, [0, 200], [30, 100], [0, 1654]);

    addDemoSamples(demoData, "/v2/", 200, 3000, [0, 100], [30, 100], [0, 1654]);
    addDemoSamples(demoData, "/v2/", 400, 500, [0, 100], [30, 100], [0, 1654]);

    addDemoSamples(demoData, "/v2/", 200, 1000, [0, 60], [30, 100], [0, 1654]);
    addDemoSamples(demoData, "/v2/", 400, 50, [0, 60], [30, 100], [0, 1654]);

    addDemoSamples(demoData, "/v2/", 200, 500, [0, 40], [30, 100], [0, 1654]);
    addDemoSamples(demoData, "/v2/", 400, 25, [0, 40], [30, 100], [0, 1654]);

    addDemoSamples(demoData, "/v2/", 200, 250, [0, 10], [30, 100], [0, 1654]);
    addDemoSamples(demoData, "/v2/", 400, 20, [0, 10], [30, 100], [0, 1654]);

    addDemoSamples(demoData, "/v2/", 200, 125, [0, 5], [30, 100], [0, 1654]);
    addDemoSamples(demoData, "/v2/", 400, 10, [0, 5], [30, 100], [0, 1654]);


    addDemoSamples(demoData, "/v2/", 200, 7000, [0, 70], [100, 200], [0, 2054]);
    addDemoSamples(demoData, "/v2/help", 400, 800, [0, 70], [100, 400], [0, 2054]);

    addDemoSamples(demoData, "/v3/", 404, 300, [0, 40], [100, 200], [0, 2054]);
    addDemoSamples(demoData, "/v3/help", 404, 53, [0, 40], [100, 400], [0, 2054]);

    addDemoSamples(demoData, "/v2/account", 200, 1000, [0, 24], [100, 200], [0, 2054]);
    addDemoSamples(demoData, "/v1/", 400, 80, [0, 24], [100, 400], [0, 2054]);

    addDemoSamples(demoData, "/v2/account", 200, 300, [122, 14], [100, 300], [0, 1654]);
    addDemoSamples(demoData, "/v2/account", 500, 800, [122, 14], [100, 300], [0, 1654]);

    addDemoSamples(demoData, "/v2/account", 200, 300, [130, 8], [100, 300], [0, 1654]);
    addDemoSamples(demoData, "/v2/account", 400, 800, [130, 8], [100, 300], [0, 1654]);

    addDemoSamples(demoData, "/v2/account", 200, 200, [135, 5], [100, 300], [0, 1654]);
    addDemoSamples(demoData, "/v2/account", 500, 700, [135, 5], [100, 300], [0, 1654]);

    addDemoSamples(demoData, "/v2/account", 200, 200, [138, 2], [100, 300], [0, 1654]);
    addDemoSamples(demoData, "/v2/account", 500, 700, [138, 2], [100, 300], [0, 1654]);

    addDemoSamples(demoData, "/v2/account", 200, 150, [139, 1], [200, 1000], [0, 1654]);
    addDemoSamples(demoData, "/v2/account", 500, 150, [139, 1], [200, 1000], [0, 1654]);

    addDemoSamples(demoData, "/v2/account", 200, 8000, [0, 140], [200, 300], [0, 1654]);
    addDemoSamples(demoData, "/v2/account", 500, 800, [0, 140], [200, 300], [0, 1654]);
    return demoData
}