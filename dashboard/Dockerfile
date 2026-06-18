# Build stage
FROM node:20-alpine AS builder

WORKDIR /app
RUN corepack enable

# Install dependencies (pnpm, per package.json "packageManager")
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# Backend URL baked into the build (see vite.config.ts / src/lib/url.ts).
ARG SERVER_URL
ENV SERVER_URL=$SERVER_URL
# Optional base path when serving the dashboard under a sub-path (e.g. /api-analytics).
ARG RELATIVE_DASHBOARD_URL
ENV VITE_RELATIVE_DASHBOARD_URL=$RELATIVE_DASHBOARD_URL

# Build a standalone Node server
ENV ADAPTER=node

COPY . .
RUN pnpm build

# Production stage
FROM node:20-alpine AS production

# needed for healthcheck
RUN apk add --no-cache curl

WORKDIR /app
RUN corepack enable

# Copy built output and install production dependencies only
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json /app/pnpm-lock.yaml ./
RUN pnpm install --prod --frozen-lockfile

# Start the server
CMD ["node", "build"]
