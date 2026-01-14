# API Analytics Tools

Collection of maintenance and monitoring tools for the API Analytics platform.

## Tools Overview

### checkup
Interactive CLI tool for viewing system health and usage statistics.

**Purpose**: Manual inspection of system status, database stats, and usage metrics

**Usage**:
```bash
./bin/checkup              # Full system checkup
./bin/checkup --users      # User statistics
./bin/checkup --monitors   # Monitor statistics
./bin/checkup --database   # Database statistics
./bin/checkup --email      # Email daily report
```

**Environment Variables**:
- `POSTGRES_URL` (required)
- `API_BASE_URL` (optional, for API tests)
- `MONITOR_API_KEY` (optional, for API tests)
- `MONITOR_USER_ID` (optional, for API tests)
- `EMAIL_ADDRESS` (optional, for email reports)

### cleanup
Scheduled maintenance tool for deleting expired data.

**Purpose**: Periodic cleanup of old requests and expired users

**Usage**:
```bash
./bin/cleanup                                    # Clean old requests
./bin/cleanup --users                            # Clean expired users
./bin/cleanup --target-user <api-key>           # Delete specific user
./bin/cleanup --requests-limit 2000000          # Set request limit per user
./bin/cleanup --user-expiry 4320h               # Set user expiry (180 days)
```

**Environment Variables**:
- `POSTGRES_URL` (required)

**Recommended Schedule**: Daily via cron

### monitor
Continuous monitoring service that checks system health and sends email alerts.

**Purpose**: Production monitoring with email notifications on failures

**Usage**:
```bash
./bin/monitor  # Run once to check system status
```

**Environment Variables**:
- `POSTGRES_URL` (required)
- `API_BASE_URL` (required)
- `MONITOR_API_KEY` (required)
- `MONITOR_USER_ID` (required)
- `EMAIL_ADDRESS` (required)

**Recommended Deployment**: Containerized service or scheduled every 5-15 minutes

## Shared Libraries

### usage
Shared library providing database query functions for usage statistics.

**Packages**:
- `usage/usage` - Core utilities and types
- `usage/users` - User-related queries
- `usage/requests` - Request-related queries
- `usage/monitors` - Monitor-related queries

### config
Shared configuration package for environment variable management.

**Usage**:
```go
import "github.com/tom-draper/api-analytics/server/tools/config"

// Load all config
cfg, err := config.Load()

// Load with required fields
cfg, err := config.LoadWithRequired("POSTGRES_URL", "API_BASE_URL")

// Create database from config
db, err := cfg.NewDatabase(ctx)
```

### monitor/pkg
Public API for monitor functionality, used by checkup tool.

## Building

Each tool has its own Makefile for building:

```bash
# Build checkup
cd checkup && make build

# Build cleanup
cd cleanup && make build

# Build monitor
cd monitor && make build
```

Clean build artifacts:
```bash
cd <tool> && make clean
```

Run tests:
```bash
cd <tool> && make test
```

Install to system:
```bash
cd <tool> && sudo make install
```

## Architecture

```
tools/
├── checkup/          # Interactive stats viewer
│   ├── Makefile      # Build checkup
│   ├── cmd/checkup/  # Main entry point
│   ├── internal/     # Display & check logic
│   └── bin/checkup   # Built binary
├── cleanup/          # Data cleanup utility
│   ├── Makefile      # Build cleanup
│   ├── cmd/cleanup/  # Main entry point
│   ├── internal/     # Cleanup logic
│   └── bin/cleanup   # Built binary
├── monitor/          # System monitor
│   ├── Makefile      # Build monitor
│   ├── cmd/monitor/  # Main entry point
│   ├── internal/     # Monitor logic
│   ├── pkg/          # Public API (used by checkup)
│   └── bin/monitor   # Built binary
├── usage/            # Shared library (single module)
│   ├── go.mod        # Module definition
│   ├── usage/        # Core utilities package
│   ├── users/        # User queries package
│   ├── requests/     # Request queries package
│   └── monitors/     # Monitor queries package
└── config/           # Shared configuration
    └── go.mod        # Module definition
```

**Note:** All modules are managed via the workspace at `server/go.work`.

## Environment Setup

Create a `.env` file in the tools directory:

```env
# Database (required for all tools)
POSTGRES_URL=postgresql://user:password@localhost:5432/database

# API Configuration (for checkup and monitor)
API_BASE_URL=https://api.apianalytics.com/
MONITOR_API_KEY=your-monitor-api-key
MONITOR_USER_ID=your-monitor-user-id

# Email (for alerts and reports)
EMAIL_ADDRESS=alerts@yourdomain.com
```

## Design Decisions

### Why Separate Tools?

1. **Different Execution Models**: Each tool has distinct runtime characteristics
   - checkup: Interactive CLI
   - cleanup: Scheduled batch job
   - monitor: Continuous service

2. **Independent Deployment**: Tools can be deployed and updated independently

3. **Clear Responsibilities**: Each tool has a single, well-defined purpose

4. **Flexible Dependencies**: Only include what's needed in each deployment

### Shared Code

Common functionality is factored into shared libraries:
- Database queries → `usage` package
- Configuration → `config` package
- Monitor functions → `monitor/pkg` package

This provides code reuse while maintaining tool independence.
