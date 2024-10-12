#!/bin/bash

if [ $# -ge 1 ]; then
    echo "API key provided: $1"
    api_key=$1
else
    echo "Testing API key generation..."

    response=$(curl -s -X GET http://localhost:3000/api/generate)

    api_key=$(echo "$response" | sed 's/\"//g')

    # Confirm successful by a 36 character response length
    if [ ${#api_key} -ne 36 ]; then
        echo -e "\e[31mFailed\e[0m"
        echo "Response: $response"
        exit 0
    fi

    echo -e "\e[32mSuccess\e[0m"
    echo "Response: $response"
fi

# Print a newline for formatting
echo ""


echo "Testing request data logging..."

response=$(curl -s -X POST  \
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
    http://localhost:8000/api/requests
)

# Confirm successful by a 201 status code
if [[ "$response" != *"201"* ]]; then
    echo -e "\e[31mFailed\e[0m"
    echo "Response: $response"
    exit 0
fi

echo -e "\e[32mSuccess\e[0m"
echo "Response: $response"

# Print a newline for formatting
echo ""


echo "Testing raw data access..."

response=$(curl -s -X GET http://localhost:3000/api/data --header "X-AUTH-TOKEN: $api_key" )

# Confirm successful by reasonable response length
if [ ${#response} -lt 100 ]; then
    echo -e "\e[31mFailed\e[0m"
    echo "Response: $response"
    exit 0
fi

echo -e "\e[32mSuccess\e[0m"
echo "Response: $response"

# Print a newline for formatting
echo ""


echo "Testing account deletion..."

# Delete API key and associated data generated in this test
response=$(curl -s -X GET http://localhost:3000/api/delete/$api_key)

# Confirm successful by a 200 status code
if [[ "$response" != *"200"* ]]; then
    echo -e "\e[31mFailed\e[0m"
    echo "Response: $response"
    exit 0
fi

echo -e "\e[32mSuccess\e[0m"
echo "Response: $response"