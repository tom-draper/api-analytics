#!/bin/bash
# Creates a .env file from .env.example with a randomly generated database password.

set -euo pipefail

BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

if [ -f .env ]; then
    echo -e "${YELLOW}A .env file already exists. Delete it first if you want to start fresh.${NC}"
    exit 1
fi

if ! command -v openssl &>/dev/null; then
    echo "openssl is required to generate a password but was not found."
    exit 1
fi

password=$(openssl rand -hex 32)
hash_secret=$(openssl rand -hex 32)

sed -e "s|POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD=$password|" \
    -e "s|USER_HASH_SECRET=.*|USER_HASH_SECRET=$hash_secret|" \
    .env.example > .env

echo -e "${GREEN}${BOLD}.env created.${NC}"
echo ""
echo "Generated a random database password and user-hash secret."
echo ""
echo -e "Next steps:"
echo -e "  1. Edit ${BOLD}.env${NC} and fill in any remaining values"
echo -e "     - For SSL setups: set DOMAIN_NAME and ACME_EMAIL"
echo -e "  2. Run: ${BOLD}docker compose up -d${NC}"
echo -e "  3. Test: ${BOLD}./tests/test-internal.sh${NC}"
