function getDemoStatus(date: Date, status: number) {
    let start = new Date()
    start.setDate(start.getDate() - 100);
    let end = new Date()
    end.setDate(end.getDate() - 96);

    if (date > start && date < end) {
        return 400
    } else {
        return status
    }
}

function getDemoUserAgent() {
    let p = Math.random()
    if (p < 0.19) {
        return "Mozilla/5.0 (iPhone; CPU iPhone OS 13_5_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.1.1 Mobile/15E148 Safari/604.1"
    } else if (p < 0.3) {
        return "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0) Gecko/20100101 Firefox/47.0"
    } else if (p < 0.34) {
        return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59"
    } else if (p < 0.36) {
        return "curl/7.64.1"
    } else if (p < 0.39) {
        return "PostmanRuntime/7.26.5"
    } else if (p < 0.4) {
        return "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.106 Safari/537.36 OPR/38.0.2220.41"
    } else if (p < 0.4) {
        return "Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:42.0) Gecko/20100101 Firefox/42.0"
    } else {
        return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36"
    }
}

function addDemoSamples(
    demoData: any[],
    endpoint: string,
    status: number,
    count: number
) {
    for (let i = 0; i < count; i++) {
        let date = new Date();
        date.setDate(date.getDate() - Math.floor(Math.random() * 650));
        date.setHours(Math.floor(Math.random() * 24));
        demoData.push({
            hostname: "demo-api.com",
            path: endpoint,
            user_agent: getDemoUserAgent(),
            method: 0,
            status: getDemoStatus(date, status),
            response_time: Math.floor(Math.random() * 240 + 55),
            created_at: date.toISOString(),
        });
    }
}

export default function genDemoData() {
    let demoData = [];

    addDemoSamples(demoData, "/v1/", 200, 18000);
    addDemoSamples(demoData, "/v1/", 400, 1000);
    addDemoSamples(demoData, "/v1/account", 200, 8000);
    addDemoSamples(demoData, "/v1/account", 400, 1200);
    addDemoSamples(demoData, "/v1/help", 200, 700);
    addDemoSamples(demoData, "/v1/help", 400, 70);
    addDemoSamples(demoData, "/v2/", 200, 34000);
    addDemoSamples(demoData, "/v2/", 400, 200);
    addDemoSamples(demoData, "/v2/account", 200, 7000);
    addDemoSamples(demoData, "/v2/account", 400, 1000);

    for (let i = 0; i < 1000; i++) {
        let date = new Date();
        date.setDate(date.getDate() - Math.floor(Math.random() * 50 + 30));
        date.setHours(Math.floor(Math.random() * 24));
        demoData.push({
            hostname: "demo-api.com",
            path: '/v2/',
            user_agent:
                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
            method: 0,
            status: 200,
            response_time: Math.floor(Math.random() * 500 + 130),
            created_at: date.toISOString(),
        });
    }

    console.log(demoData);

    return demoData
}