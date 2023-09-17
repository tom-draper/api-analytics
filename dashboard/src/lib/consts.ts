export const SERVER_URL: string = "https://www.apianalytics-server.com"

// Dashboard logged requests data is delivered as an array per request rather 
// than an object to reduce payload size. The index of each value is stored as 
// a constant for readability when accessing the data.
export const IP_ADDRESS = 0;
export const PATH = 1;
export const USER_AGENT = 2;
export const METHOD = 3;
export const RESPONSE_TIME = 4;
export const STATUS = 5;
export const LOCATION = 6;
export const CREATED_AT = 7;
