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

# Check if a domain name and/or API key have been provided
if [ $# -ge 1 ]; then
    domain=$1
    echo "Domain provided: $domain"
else
    echo "Please provide the domain name as the first argument."
    exit 1
fi

if [ $# -ge 2 ]; then
    echo "API key provided: $2"
    echo ""
    api_key=$2
else
    echo "Testing external API key generation via Nginx + HTTPS..."

    response=$(curl -s -X GET -w "%{http_code}" https://$domain/api/generate)
    status_code="${response: -3}"   # Extract the last 3 characters as the status code
    response_body="${response%???}"
    api_key=$(echo "$response_body" | sed 's/[0-9][0-9][0-9]$//' | sed 's/\"//g')  # Strip status code from response

    # Check if the status code is 200 and the API key is the correct length
    display_result "$response_body" "$status_code" 200
    if [ "$status_code" -ne 200 ]; then
        exit 1
    fi
fi

echo "Testing external request data logging via Nginx + HTTPS..."

response=$(curl -so - -w "%{http_code}" -X POST  \
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
    "https://$domain/api/log-request")

status_code="${response: -3}"
response_body="${response%???}"

# Check if the status code is 201 (Created)
display_result "$response_body" "$status_code" 201
if [ "$status_code" -ne 201 ]; then
    exit 2
fi

echo "Testing external user ID access via Nginx + HTTPS..."

response=$(curl -s -X GET -w "%{http_code}" https://$domain/api/user-id/$api_key)
echo $response
status_code="${response: -3}"   # Extract the last 3 characters as the status code
response_body="${response%???}"
user_id=$(echo "$response_body" | sed 's/[0-9][0-9][0-9]$//' | sed 's/\"//g')  # Strip status code from response

# Check if the status code is 200 and the API key is the correct length
display_result "$response_body" "$status_code" 200
if [ "$status_code" -ne 200 ]; then
    exit 3
fi

echo "Testing external dashboard data access via Nginx + HTTPS..."

response=$(curl -s -X GET -w "%{http_code}" --compressed -H "Accept-Encoding: gzip" \
    https://$domain/api/requests/$user_id)

status_code="${response: -3}"   # Extract the last 3 characters as the status code
response_body="${response%???}"  # Extract everything except the last 3 characters as the response body

# Check if the status code is 200 (OK) and response has a reasonable length
display_result "$response_body" "$status_code" 200
if [ "$status_code" -ne 200 ] || [ ${#response_body} -lt 100 ]; then
    exit 4
fi

echo "Testing external raw data access via Nginx + HTTPS..."

response=$(curl -s -X GET -w "%{http_code}" \
    https://$domain/api/data --header "X-AUTH-TOKEN: $api_key")

status_code="${response: -3}"   # Extract the last 3 characters as the status code
response_body="${response%???}"  # Extract everything except the last 3 characters as the response body

# Check if the status code is 200 (OK) and response has a reasonable length
display_result "$response_body" "$status_code" 200
if [ "$status_code" -ne 200 ] || [ ${#response_body} -lt 100 ]; then
    exit 5
fi

echo "Testing external account deletion via Nginx + HTTPS..."

response=$(curl -s -X GET -w "%{http_code}" https://$domain/api/delete/$api_key)

status_code="${response: -3}"   # Extract the last 3 characters as the status code
response_body="${response%???}"  # Extract everything except the last 3 characters as the response body

# Check if the status code is 200 (OK)
display_result "$response_body" "$status_code" 200
if [ "$status_code" -ne 200 ]; then
    exit 6
fi