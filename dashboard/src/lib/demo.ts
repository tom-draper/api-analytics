function randomChoice(p: any[]) {
    let rnd = p.reduce((a, b) => a + b) * Math.random();
    return p.findIndex((a) => (rnd -= a) < 0);
}

function randomChoices(p: number[], count: number): number[] {
    return Array.from(Array(count), randomChoice.bind(null, p));
}

const userAgents = [
    "Mozilla/5.0 (iPhone; CPU iPhone OS 13_5_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.1.1 Mobile/15E148 Safari/604.1",
    "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0) Gecko/20100101 Firefox/47.0",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59",
    "curl/7.64.1",
    "PostmanRuntime/7.26.5",
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.106 Safari/537.36 OPR/38.0.2220.41",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:42.0) Gecko/20100101 Firefox/42.0",
    "python-requests/2.26.0",
];
const userAgentDist = [0.19, 0.11, 0.04, 0.01, 0.01, 0.01, 0.01, 0.01];

function getUserAgent() {
    return randInt(0, userAgentDist.length);
}

function getUserAgents() {
    const userAgentsMapping: UserAgents = {};
    for (let i = 0; i < 100; i++) {
        userAgentsMapping[i] = userAgents[i];
    }
    return userAgentsMapping;
}

const hoursDist = (() => {
    const p = Array(24).fill(1);
    // Create daytime peak
    for (let i = 0; i < 3; i++) {
        for (let j = 5 + i; j < 11 - i; j++) {
            p[j] += 0.15;
        }
    }
    // Create evening peak
    for (let i = 0; i < 4; i++) {
        for (let j = 11 + i; j < 21 - i; j++) {
            p[j] += 0.15;
        }
    }
    return p;
})();

function getDemoHour() {
    return randomChoices(hoursDist, 1)[0];
}

const demoLocations = [
    "GB",
    "US",
    "MX",
    "CA",
    "FR",
    "AU",
    "ID",
    "IE",
    "DE",
    "PL",
    "AF",
    "AL",
    "DZ",
    "AS",
    "AD",
    "AO",
    "AI",
    "AQ",
    "AG",
    "AR",
    "AM",
    "AW",
    "AT",
    "AZ",
    "BS",
    "BH",
    "BD",
    "BB",
    "BY",
    "BE",
    "BZ",
    "BJ",
    "BM",
    "BT",
    "BO",
    "BQ",
];
const demoLocationsDist = [
    0.56, 1, 0.18, 0.2, 0.4, 0.3, 0.1, 0.05, 0.2, 0.06, 0.01, 0.01, 0.01, 0.01,
    0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01,
    0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01,
];

function getLocation() {
    const idx = randomChoice(demoLocationsDist);
    return demoLocations[idx];
}

const demoHostnames = ["example.com", "example2.com", "example3.com"];
const demoHostnamesDist = [0.5, 0.35, 0.15];

function getHostname() {
    const idx = randomChoice(demoHostnamesDist);
    return demoHostnames[idx];
}

function daysToSeconds(days: number) {
    return days * 24 * 60 * 60;
}

function getDate(daysAgo: Range, distribution: Distribution, outages: Range[]) {
    let secondsAgo: number;
    switch (distribution) {
        case Distribution.Uniform:
            secondsAgo = randInt(
                daysToSeconds(daysAgo.min),
                daysToSeconds(daysAgo.max)
            );
            break;
        case Distribution.Poisson:
            secondsAgo = samplePoisson(3, daysToSeconds(daysAgo.max));
            break;
        case Distribution.Normal:
            secondsAgo = randn_bm(
                daysToSeconds(daysAgo.min),
                daysToSeconds(daysAgo.max),
                0.75
            );
            break;
    }

    // Check if chosen date falls inside of an outage
    for (let outage of outages) {
        if (secondsAgo > daysToSeconds(outage.min) && secondsAgo < daysToSeconds(outage.max)) {
            return null
        }
    }

    const now = new Date();
    const date = new Date(now.getTime() - secondsAgo * 1000);
    date.setHours(getDemoHour(), Math.floor(Math.random() * 60))
    return date;
}

function randInt(min: number, max: number) {
    return Math.floor(Math.random() * max + min);
}

type Sample = {
    count: number;
    endpoint: string;
    status: number;
    method?: number;
    daysAgo: Range;
    responseTime: Range;
    user: Range;
    distribution: Distribution;
};

type Range = {
    min: number;
    max: number;
};

function addDemoSamples(demoData: RequestsData, config: Sample, outages: Range[] = []) {
    for (let i = 0; i < config.count; i++) {
        const date = getDate(config.daysAgo, config.distribution, outages);
        if (date === null) {
            continue
        }
        demoData.push([
            randInt(config.user.min, config.user.max).toString(),
            config.endpoint,
            getHostname(),
            getUserAgent(),
            config.method || 0,
            randInt(config.responseTime.min, config.responseTime.max),
            config.status,
            getLocation(),
            null,
            date
        ]);
    }
}

const enum Distribution {
    Uniform,
    Normal,
    Poisson,
}

function samplePoisson(lambda: number, T: number) {
    const sampleNextEventTime = () => {
        return -Math.log(Math.random()) / lambda;
    };

    let time = 0;
    let nEvents = 0;
    while (time < T) {
        time += sampleNextEventTime();
        nEvents++;
    }
    return nEvents - 1;
}

function gaussianRand() {
    let rand = 0;
    for (let i = 0; i < 6; i += 1) {
        rand += Math.random();
    }
    return rand / 6;
}

function randn_bm(min, max, skew) {
    let u = 0, v = 0;
    while (u === 0) u = Math.random() //Converting [0,1) to (0,1)
    while (v === 0) v = Math.random()
    let num = Math.sqrt(-2.0 * Math.log(u)) * Math.cos(2.0 * Math.PI * v)

    num = num / 10.0 + 0.5 // Translate to 0 -> 1
    if (num > 1 || num < 0)
        num = randn_bm(min, max, skew) // resample between 0 and 1 if out of range

    else {
        num = Math.pow(num, skew) // Skew
        num *= max - min // Stretch to fill range
        num += min // offset to min
    }
    return num
}

function gaussianRandom(start: number, end: number) {
    return Math.floor(start + gaussianRand() * (end - start + 1));
}

function scaleRange(range: Range, scale: number): Range {
    return {
        min: range.min * scale,
        max: range.max * scale,
    };
}

function createUniformBaselineSamples(
    demoRequests: RequestsData,
    maxDaysAgo: number,
    scale: number,
    outages: Range[]
) {
    const count = 12000 * scale;

    const maxRange = { min: 0, max: maxDaysAgo }
    const v1ResponseTime = { min: 35, max: 163 };
    const v2ResponseTime = { min: 19, max: 137 };
    const v3ResponseTime = { min: 10, max: 80 };
    const notFoundResponseTime = { min: 6, max: 30 };
    const user = { min: 0, max: count * 0.006 }

    addDemoSamples(demoRequests, {
        count: count,
        endpoint: "/api/v1/",
        status: 200,
        daysAgo: maxRange,
        responseTime: v1ResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.001,
        endpoint: "/api/v1/",
        status: 400,
        daysAgo: maxRange,
        responseTime: v1ResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.01,
        endpoint: "/api/v1/",
        status: 500,
        daysAgo: maxRange,
        responseTime: v1ResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.45,
        endpoint: "/api/v1/account",
        status: 200,
        daysAgo: maxRange,
        responseTime: v1ResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.011,
        endpoint: "/api/v1/account",
        status: 400,
        daysAgo: maxRange,
        responseTime: scaleRange(v1ResponseTime, 4),
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.21,
        endpoint: "/api/v1/account",
        status: 500,
        daysAgo: maxRange,
        responseTime: scaleRange(v1ResponseTime, 4),
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.07,
        endpoint: "/api/v1/account",
        status: 504,
        daysAgo: maxRange,
        responseTime: scaleRange(v1ResponseTime, 4),
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.0053,
        endpoint: "/api/v1/accounts",
        status: 404,
        daysAgo: maxRange,
        responseTime: notFoundResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);


    addDemoSamples(demoRequests, {
        count: count * 1.42,
        endpoint: "/api/v2/",
        status: 200,
        daysAgo: maxRange,
        responseTime: v2ResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.0009,
        endpoint: "/api/v2/",
        status: 400,
        daysAgo: maxRange,
        responseTime: v2ResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.33,
        endpoint: "/api/v2/account",
        status: 200,
        daysAgo: maxRange,
        responseTime: v2ResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.023,
        endpoint: "/api/v2/account",
        status: 400,
        daysAgo: maxRange,
        responseTime: v2ResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.0051,
        endpoint: "/api/v2/account",
        status: 500,
        daysAgo: maxRange,
        responseTime: v2ResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.007,
        endpoint: "/api/v2/account",
        status: 504,
        daysAgo: maxRange,
        responseTime: v2ResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.33,
        endpoint: "/api/v2/help",
        status: 200,
        daysAgo: maxRange,
        responseTime: v2ResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.2,
        endpoint: "/api/v2/info",
        status: 200,
        daysAgo: maxRange,
        responseTime: v2ResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.62,
        endpoint: "/api/v2/account",
        method: 1,
        status: 201,
        daysAgo: maxRange,
        responseTime: scaleRange(v3ResponseTime, 6),
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.32,
        endpoint: "/api/v2/account/delete",
        method: 1,
        status: 200,
        daysAgo: maxRange,
        responseTime: scaleRange(v3ResponseTime, 6),
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.012,
        endpoint: "/api/v2/account/delete",
        method: 1,
        status: 409,
        daysAgo: maxRange,
        responseTime: scaleRange(v3ResponseTime, 6),
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.019,
        endpoint: "/api/v2/account/delete",
        method: 1,
        status: 400,
        daysAgo: maxRange,
        responseTime: scaleRange(v3ResponseTime, 6),
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.00051,
        endpoint: "/api/v2/account",
        status: 400,
        daysAgo: maxRange,
        responseTime: v2ResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.0092,
        endpoint: "/api/v2/accounts",
        status: 404,
        daysAgo: maxRange,
        responseTime: notFoundResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);

    addDemoSamples(demoRequests, {
        count: count * 0.412,
        endpoint: "/api/v3/",
        status: 200,
        daysAgo: maxRange,
        responseTime: v3ResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.0081,
        endpoint: "/api/v3/",
        status: 400,
        daysAgo: maxRange,
        responseTime: v3ResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.052,
        endpoint: "/api/v3/account",
        status: 200,
        daysAgo: maxRange,
        responseTime: v3ResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.0022,
        endpoint: "/api/v3/account",
        status: 400,
        daysAgo: maxRange,
        responseTime: scaleRange(v3ResponseTime, 6),
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.22,
        endpoint: "/api/v3/account",
        method: 1,
        status: 201,
        daysAgo: maxRange,
        responseTime: scaleRange(v3ResponseTime, 6),
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.12,
        endpoint: "/api/v3/account/delete",
        method: 1,
        status: 200,
        daysAgo: maxRange,
        responseTime: v3ResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.00011,
        endpoint: "/api/v3/accounts",
        status: 404,
        daysAgo: maxRange,
        responseTime: notFoundResponseTime,
        user: user,
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.0016,
        endpoint: "/api/v3/admin",
        status: 403,
        daysAgo: {
            min: 0,
            max: maxDaysAgo / 5,
        },
        responseTime: {
            min: 25,
            max: 600,
        },
        user: {
            min: 115,
            max: 130,
        },
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.0026,
        endpoint: "/api/v3/admin",
        status: 200,
        daysAgo: {
            min: 0,
            max: maxDaysAgo / 5,
        },
        responseTime: {
            min: 25,
            max: 600,
        },
        user: {
            min: 115,
            max: 120,
        },
        distribution: Distribution.Uniform,
    }, outages);
}

function createVariableBaselineSamples(
    demoRequests: RequestsData,
    maxDaysAgo: number,
    scale: number,
    outages: Range[]
) {
    const count = 2000 * scale;
    addDemoSamples(demoRequests, {
        count: count,
        endpoint: "/api/v1/",
        status: 200,
        daysAgo: {
            min: 0,
            max: maxDaysAgo,
        },
        responseTime: {
            min: 25,
            max: 600,
        },
        user: {
            min: 115,
            max: 390,
        },
        distribution: Distribution.Normal,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count,
        endpoint: "/api/v3/",
        status: 200,
        daysAgo: {
            min: 0,
            max: maxDaysAgo / 2,
        },
        responseTime: {
            min: 25,
            max: 600,
        },
        user: {
            min: 115,
            max: 390,
        },
        distribution: Distribution.Normal,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count,
        endpoint: "/api/v2/",
        status: 200,
        daysAgo: {
            min: 0,
            max: maxDaysAgo / 3,
        },
        responseTime: {
            min: 25,
            max: 600,
        },
        user: {
            min: 115,
            max: 390,
        },
        distribution: Distribution.Normal,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count,
        endpoint: "/api/v1/account",
        status: 200,
        daysAgo: {
            min: 0,
            max: maxDaysAgo / 4,
        },
        responseTime: {
            min: 25,
            max: 600,
        },
        user: {
            min: 115,
            max: 390,
        },
        distribution: Distribution.Normal,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 2,
        endpoint: "/api/v2/account",
        status: 200,
        daysAgo: {
            min: 0,
            max: maxDaysAgo / 5,
        },
        responseTime: {
            min: 25,
            max: 600,
        },
        user: {
            min: 115,
            max: 390,
        },
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 2,
        endpoint: "/api/v3/account",
        status: 200,
        daysAgo: {
            min: 0,
            max: maxDaysAgo / 5,
        },
        responseTime: {
            min: 25,
            max: 600,
        },
        user: {
            min: 115,
            max: 390,
        },
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.5,
        endpoint: "/api/v1/account",
        status: 200,
        daysAgo: {
            min: 0,
            max: maxDaysAgo / 4,
        },
        responseTime: {
            min: 25,
            max: 600,
        },
        user: {
            min: 115,
            max: 390,
        },
        distribution: Distribution.Uniform,
    }, outages);

}

function createVariableUsageSamples(
    demoRequests: RequestsData,
    maxDaysAgo: number,
    scale: number,
    outages: Range[]
) {
    const count = 1000 * scale;
    addDemoSamples(demoRequests, {
        count: count,
        endpoint: "/api/v1/",
        status: 200,
        daysAgo: {
            min: maxDaysAgo * 0.7,
            max: maxDaysAgo * 0.75,
        },
        responseTime: {
            min: 60,
            max: 600,
        },
        user: {
            min: 115,
            max: 116,
        },
        distribution: Distribution.Normal,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.3,
        endpoint: "/api/v1/",
        status: 200,
        daysAgo: {
            min: maxDaysAgo * 0.79,
            max: maxDaysAgo * 0.83,
        },
        responseTime: {
            min: 60,
            max: 600,
        },
        user: {
            min: 116,
            max: 120,
        },
        distribution: Distribution.Normal,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count,
        endpoint: "/api/v3/",
        status: 200,
        daysAgo: {
            min: maxDaysAgo * 0.7,
            max: maxDaysAgo * 0.9,
        },
        responseTime: {
            min: 60,
            max: 600,
        },
        user: {
            min: 116,
            max: 120,
        },
        distribution: Distribution.Normal,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 2,
        endpoint: "/api/v2/",
        status: 200,
        daysAgo: {
            min: maxDaysAgo * 0.3,
            max: maxDaysAgo * 0.6,
        },
        responseTime: {
            min: 60,
            max: 600,
        },
        user: {
            min: 116,
            max: 120,
        },
        distribution: Distribution.Normal,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 2,
        endpoint: "/api/v2/",
        status: 200,
        daysAgo: {
            min: maxDaysAgo * 0.3,
            max: maxDaysAgo * 0.6,
        },
        responseTime: {
            min: 60,
            max: 600,
        },
        user: {
            min: 116,
            max: 120,
        },
        distribution: Distribution.Normal,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 2,
        endpoint: "/api/v2/",
        status: 200,
        daysAgo: {
            min: maxDaysAgo * 0.1,
            max: maxDaysAgo * 0.55,
        },
        responseTime: {
            min: 60,
            max: 600,
        },
        user: {
            min: 116,
            max: 120,
        },
        distribution: Distribution.Normal,
    }, outages);

    addDemoSamples(demoRequests, {
        count: count * 2,
        endpoint: "/api/v2/account",
        status: 200,
        daysAgo: {
            min: maxDaysAgo * 0.1,
            max: maxDaysAgo * 0.55,
        },
        responseTime: {
            min: 60,
            max: 600,
        },
        user: {
            min: 116,
            max: 120,
        },
        distribution: Distribution.Normal,
    }, outages);
}

function createErrorSamples(
    demoRequests: RequestsData,
    maxDaysAgo: number,
    scale: number,
    outages: Range[]
) {
    const count = 300 * scale;
    addDemoSamples(demoRequests, {
        count: count,
        endpoint: "/api/v1/",
        status: 500,
        daysAgo: {
            min: 200,
            max: 210
        },
        responseTime: {
            min: 100,
            max: 600,
        },
        user: {
            min: 116,
            max: 120,
        },
        distribution: Distribution.Normal,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.6,
        endpoint: "/api/v1/account",
        status: 500,
        daysAgo: {
            min: 250,
            max: 260
        },
        responseTime: {
            min: 100,
            max: 600,
        },
        user: {
            min: 116,
            max: 120,
        },
        distribution: Distribution.Normal,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.6,
        endpoint: "/api/v2/account",
        status: 500,
        daysAgo: {
            min: 250,
            max: 260
        },
        responseTime: {
            min: 100,
            max: 600,
        },
        user: {
            min: 116,
            max: 120,
        },
        distribution: Distribution.Normal,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.6,
        endpoint: "/api/v4/account",
        status: 404,
        daysAgo: {
            min: 24,
            max: 27
        },
        responseTime: {
            min: 100,
            max: 600,
        },
        user: {
            min: 116,
            max: 120,
        },
        distribution: Distribution.Normal,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.6,
        endpoint: "/api/v3/account/test",
        status: 404,
        daysAgo: {
            min: 20,
            max: 40
        },
        responseTime: {
            min: 100,
            max: 600,
        },
        user: {
            min: 116,
            max: 120,
        },
        distribution: Distribution.Normal,
    }, outages);

    addDemoSamples(demoRequests, {
        count: count,
        endpoint: "/api/robots.txt",
        status: 404,
        daysAgo: {
            min: 0,
            max: maxDaysAgo
        },
        responseTime: {
            min: 10,
            max: 200,
        },
        user: {
            min: 100,
            max: 120,
        },
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 3,
        endpoint: "/robots.txt",
        status: 404,
        daysAgo: {
            min: 0,
            max: maxDaysAgo
        },
        responseTime: {
            min: 10,
            max: 20,
        },
        user: {
            min: 100,
            max: 120,
        },
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count,
        endpoint: "/api/.env",
        status: 404,
        daysAgo: {
            min: 0,
            max: maxDaysAgo
        },
        responseTime: {
            min: 10,
            max: 20,
        },
        user: {
            min: 100,
            max: 120,
        },
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count,
        endpoint: "/.env",
        status: 404,
        daysAgo: {
            min: 0,
            max: maxDaysAgo
        },
        responseTime: {
            min: 10,
            max: 20,
        },
        user: {
            min: 100,
            max: 120,
        },
        distribution: Distribution.Uniform,
    }, outages);
    addDemoSamples(demoRequests, {
        count: count * 0.05,
        endpoint: "/test",
        status: 404,
        daysAgo: {
            min: 0,
            max: maxDaysAgo
        },
        responseTime: {
            min: 10,
            max: 20,
        },
        user: {
            min: 96,
            max: 100,
        },
        distribution: Distribution.Normal,
    }, outages);
}

export default function genDemoData() {
    const demoRequests: RequestsData = [];

    const maxDaysAgo = 460;
    const scale = 1.5;

    const outages = [{ min: 300, max: 305 }, { min: 310, max: 313 }]

    // Baseline
    createUniformBaselineSamples(demoRequests, maxDaysAgo, scale, outages);
    createVariableBaselineSamples(demoRequests, maxDaysAgo, scale, outages);
    // Usage
    createVariableUsageSamples(demoRequests, maxDaysAgo, scale, outages);
    // Anomalies
    createErrorSamples(demoRequests, maxDaysAgo, scale, outages);

    const demoData: DashboardData = {
        user_agents: getUserAgents(),
        requests: demoRequests,
    };
    return demoData;
}