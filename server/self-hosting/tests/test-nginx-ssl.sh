#!/bin/bash

display_result() {
    local response=$1
    local status_code=$2
    local expected_code=$3

    if [ "$status_code" -eq "$expected_code" ]; then
        echo -e "\e[32mSuccess\e[0m"  # Green "Success" message
    else
        echo -e "\e[31mFailed\e[0m"  # Red "Failed" message
    fi

    echo "Status Code: $status_code"
    echo "Response: $response"
    echo ""  # Print a new line for formatting
}

# Check if an API key has been provided
if [ $# -ge 1 ]; then
    echo "API key provided: $1"
    echo ""
    api_key=$1
else
    echo "Testing internal API key generation via Nginx + HTTPS..."

    response=$(curl -sk -X GET -w "%{http_code}" https://localhost/api/generate)
    status_code="${response: -3}"   # Extract the last 3 characters as the status code
    response_body="${response%???}"
    api_key=$(echo "$response_body" | sed 's/[0-9][0-9][0-9]$//' | sed 's/\"//g')  # Strip status code from response

    # Check if the status code is 200 and the API key is the correct length
    display_result "$response_body" "$status_code" 200
    if [ "$status_code" -ne 200 ]; then
        exit 1
    fi
fi

echo "Testing internal request data logging via Nginx + HTTPS..."

response=$(curl -sko - -w "%{http_code}" -X POST  \
    -H "Content-Type: application/json" \
    -d '{
        "api_key": "'"$api_key"'",
        "requests": [
            {
                "hostname": "example.com",
                "ip_address": "",
                "user_agent": "test",
                "path": "/test",
                "status": 200,
                "method": "GET",
                "response_time": 15,
                "user_id": "",
                "created_at": "'"$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")"'"
            },
            {
                "hostname": "example.com",
                "ip_address": "",
                "user_agent": "test",
                "path": "/not-found",
                "status": 404,
                "method": "GET",
                "response_time": 10,
                "user_id": "",
                "created_at": "'"$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")"'"
            }
        ],
        "framework": "FastAPI",
        "privacy_level": 0
    }' \
    https://localhost/api/log-request)

status_code="${response: -3}"
response_body="${response%???}"

# Check if the status code is 201 (Created)
display_result "$response_body" "$status_code" 201
if [ "$status_code" -ne 201 ]; then
    exit 2
fi

echo "Testing internal user ID access via Nginx + HTTPS..."

response=$(curl -sk -X GET -w "%{http_code}" https://localhost/api/user-id/$api_key)
echo $response
status_code="${response: -3}"   # Extract the last 3 characters as the status code
response_body="${response%???}"
user_id=$(echo "$response_body" | sed 's/[0-9][0-9][0-9]$//' | sed 's/\"//g')  # Strip status code from response

# Check if the status code is 200 and the API key is the correct length
display_result "$response_body" "$status_code" 200
if [ "$status_code" -ne 200 ]; then
    exit 3
fi

echo "Testing internal dashboard data access via Nginx + HTTPS..."

response=$(curl -sk -X GET -w "%{http_code}" --compressed -H "Accept-Encoding: gzip" \
    https://localhost/api/requests/$user_id)

status_code="${response: -3}"   # Extract the last 3 characters as the status code
response_body="${response%???}"  # Extract everything except the last 3 characters as the response body

# Check if the status code is 200 (OK) and response has a reasonable length
display_result "$response_body" "$status_code" 200
if [ "$status_code" -ne 200 ] || [ ${#response_body} -lt 100 ]; then
    exit 4
fi

echo "Testing internal raw data access via Nginx + HTTPS..."

response=$(curl -sk -X GET -w "%{http_code}" \
    https://localhost/api/data --header "X-AUTH-TOKEN: $api_key")

status_code="${response: -3}"   # Extract the last 3 characters as the status code
response_body="${response%???}"  # Extract everything except the last 3 characters as the response body

# Check if the status code is 200 (OK) and response has a reasonable length
display_result "$response_body" "$status_code" 200
if [ "$status_code" -ne 200 ] || [ ${#response_body} -lt 100 ]; then
    exit 5
fi

echo "Testing internal account deletion via Nginx + HTTPS..."

response=$(curl -sk -X GET -w "%{http_code}" https://localhost/api/delete/$api_key)

status_code="${response: -3}"   # Extract the last 3 characters as the status code
response_body="${response%???}"  # Extract everything except the last 3 characters as the response body

# Check if the status code is 200 (OK)
display_result "$response_body" "$status_code" 200
if [ "$status_code" -ne 200 ]; then
    exit 6
fi