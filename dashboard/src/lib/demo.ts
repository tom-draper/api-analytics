
function getDemoStatus(date: Date, status: number) {
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

function getDemoUserAgent() {
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
    } else if (p < 0.4) {
        return "Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:42.0) Gecko/20100101 Firefox/42.0";
    } else {
        return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36";
    }
}

function randomChoice(p: number[]): number {
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

function addDemoSamples(
    demoData: any[],
    endpoint: string,
    status: number,
    count: number,
    maxDaysAgo: number,
    minDaysAgo: number,
    maxResponseTime: number,
    minResponseTime: number
) {
    for (let i = 0; i < count; i++) {
        let date = new Date();
        date.setDate(
            date.getDate() - Math.floor(Math.random() * maxDaysAgo + minDaysAgo)
        );
        date.setHours(getHour());
        demoData.push({
            hostname: "demo-api.com",
            path: endpoint,
            user_agent: getDemoUserAgent(),
            method: 0,
            status: getDemoStatus(date, status),
            response_time: Math.floor(
                Math.random() * maxResponseTime + minResponseTime
            ),
            created_at: date.toISOString(),
        });
    }
}

export default function genDemoData() {
    let demoData = [];
    
    addDemoSamples(demoData, "/v1/", 200, 18000, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v1/", 400, 1000, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v1/account", 200, 8000, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v1/account", 400, 1200, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v1/help", 200, 700, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v1/help", 400, 70, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v2/", 200, 35000, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v2/", 400, 200, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v2/account", 200, 14000, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v2/account", 400, 3000, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v2/account/update", 200, 6000, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v2/account/update", 400, 400, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v2/help", 200, 6000, 650, 0, 240, 55);
    addDemoSamples(demoData, "/v2/help", 400, 400, 650, 0, 240, 55);

    addDemoSamples(demoData, "/v2/account", 200, 16000, 450, 0, 100, 30);
    addDemoSamples(demoData, "/v2/account", 400, 2000, 450, 0, 100, 30);

    addDemoSamples(demoData, "/v2/help", 200, 8000, 300, 0, 100, 30);
    addDemoSamples(demoData, "/v2/help", 400, 800, 300, 0, 100, 30);

    addDemoSamples(demoData, "/v2/", 200, 4000, 200, 0, 100, 30);
    addDemoSamples(demoData, "/v2/", 400, 800, 200, 0, 100, 30);

    addDemoSamples(demoData, "/v2/", 200, 3000, 100, 0, 100, 30);
    addDemoSamples(demoData, "/v2/", 400, 500, 100, 0, 100, 30);

    addDemoSamples(demoData, "/v2/", 200, 1000, 60, 0, 100, 30);
    addDemoSamples(demoData, "/v2/", 400, 50, 60, 0, 100, 30);

    addDemoSamples(demoData, "/v2/", 200, 500, 40, 0, 100, 30);
    addDemoSamples(demoData, "/v2/", 400, 25, 40, 0, 100, 30);

    addDemoSamples(demoData, "/v2/", 200, 250, 10, 0, 100, 30);
    addDemoSamples(demoData, "/v2/", 400, 20, 10, 0, 100, 30);

    addDemoSamples(demoData, "/v2/", 200, 125, 5, 0, 100, 30);
    addDemoSamples(demoData, "/v2/", 400, 10, 5, 0, 100, 30);

    return demoData
}