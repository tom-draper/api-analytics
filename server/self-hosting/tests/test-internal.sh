#!/bin/bash
# Tests API and logger services directly, bypassing the reverse proxy.
# Run this first to confirm the core services are working before testing your proxy setup.

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

API_URL="http://localhost:3000"
LOGGER_URL="http://localhost:8000"

passed=0
failed=0

result() {
    local label="$1"
    local status_code="$2"
    local expected="$3"
    local hint="$4"
    local body="$5"

    printf "  %-45s" "$label"
    if [ "$status_code" -eq "$expected" ]; then
        echo -e "${GREEN}PASS${NC}"
        ((passed++))
        return 0
    else
        echo -e "${RED}FAIL${NC} (HTTP $status_code)"
        echo -e "    ${YELLOW}Hint: $hint${NC}"
        if [ -n "$body" ]; then
            echo "    Response: $body"
        fi
        ((failed++))
        return 1
    fi
}

echo ""
echo -e "${BOLD}API Analytics - Internal Service Tests${NC}"
echo -e "API:    $API_URL"
echo -e "Logger: $LOGGER_URL"
echo ""

# --- API key generation ---

response=$(curl -sf -w "%{http_code}" "$API_URL/api/generate" 2>/dev/null || echo "000")
status_code="${response: -3}"
body="${response%???}"
api_key=$(echo "$body" | tr -d '"')

if ! result "API key generation" "$status_code" 200 \
    "check: docker logs api-analytics-api" "$body"; then
    echo ""
    echo -e "${RED}Cannot continue without a valid API key.${NC}"
    exit 1
fi

# --- Request logging ---

response=$(curl -sf -w "%{http_code}" -X POST \
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
    "$LOGGER_URL/api/log-request" 2>/dev/null || echo "000")

status_code="${response: -3}"
body="${response%???}"

result "Request logging" "$status_code" 201 \
    "check: docker logs api-analytics-logger" "$body" || true

# --- User ID lookup ---

response=$(curl -sf -w "%{http_code}" "$API_URL/api/user-id/$api_key" 2>/dev/null || echo "000")
status_code="${response: -3}"
body="${response%???}"
user_id=$(echo "$body" | tr -d '"')

if ! result "User ID lookup" "$status_code" 200 \
    "check: docker logs api-analytics-api" "$body"; then
    echo ""
    echo -e "${RED}Cannot continue without a valid user ID.${NC}"
    # Clean up
    curl -sf "$API_URL/api/delete/$api_key" > /dev/null 2>&1 || true
    exit 1
fi

# --- Dashboard data access ---

response=$(curl -sf -w "%{http_code}" --compressed -H "Accept-Encoding: gzip" \
    "$API_URL/api/requests/$user_id" 2>/dev/null || echo "000")
status_code="${response: -3}"
body="${response%???}"

result "Dashboard data access" "$status_code" 200 \
    "check: docker logs api-analytics-api" "$body" || true

# --- Raw data access ---

response=$(curl -sf -w "%{http_code}" \
    -H "X-AUTH-TOKEN: $api_key" \
    "$API_URL/api/data" 2>/dev/null || echo "000")
status_code="${response: -3}"
body="${response%???}"

result "Raw data access" "$status_code" 200 \
    "check: docker logs api-analytics-api" "$body" || true

# --- Account deletion ---

response=$(curl -sf -w "%{http_code}" "$API_URL/api/delete/$api_key" 2>/dev/null || echo "000")
status_code="${response: -3}"
body="${response%???}"

result "Account deletion" "$status_code" 200 \
    "check: docker logs api-analytics-api" "$body" || true

# --- Summary ---

total=$((passed + failed))
echo ""
if [ "$failed" -eq 0 ]; then
    echo -e "${GREEN}${BOLD}All $total tests passed.${NC}"
else
    echo -e "${RED}${BOLD}$failed/$total tests failed.${NC}"
    echo ""
    echo "Troubleshooting:"
    echo "  docker ps                          check all services are running"
    echo "  docker logs api-analytics-api      API service logs"
    echo "  docker logs api-analytics-logger   logger service logs"
    exit 1
fi
