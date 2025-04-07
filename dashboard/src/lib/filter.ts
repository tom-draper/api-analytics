import { ColumnIndex } from "./consts";

export type Filter = {
    timespan: [number, number];
    status: {
        success: boolean;
        redirect: boolean;
        client: boolean;
        server: boolean;
    }
    methods: {
        [method: string]: boolean;
    }
    hostnames: {
        [hostname: string]: boolean;
    }
    responseTime: [number, number];
    paths: Set<string>;
};


export function defaultFilter(data: RequestsData) {
    const filter: Filter = {
        timespan: [
            data[0][ColumnIndex.CreatedAt].getTime(),
            data[data.length - 1][ColumnIndex.CreatedAt].getTime()
        ],
        status: {
            success: true,
            redirect: true,
            client: true,
            server: true
        },
        responseTime: [0, Infinity],
        methods: getMethods(data),
        hostnames: getHostnames(data),
        paths: new Set()
    };

    return filter;
}

function getMethods(data: RequestsData) {
    const methods: { [method: string]: boolean } = {};
    for (const row of data) {
        methods[row[ColumnIndex.Method]] = true;
    }

    return methods;
}

function getHostnames(data: RequestsData) {
    const hostnames: { [hostname: string]: boolean } = {};
    for (const row of data) {
        hostnames[row[ColumnIndex.Hostname]] = true;
    }

    return hostnames;
}
