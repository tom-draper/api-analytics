# Build stage
FROM node:24-alpine AS builder

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
FROM node:24-alpine AS production

RUN adduser -D -s /bin/sh -u 1001 appuser

WORKDIR /app
RUN corepack enable

# Copy built output and install production dependencies only
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json /app/pnpm-lock.yaml ./
RUN pnpm install --prod --frozen-lockfile
RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1

CMD ["node", "build"]
