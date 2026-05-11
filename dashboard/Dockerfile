# Build stage
FROM node:24-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci
ENV PATH="/app/node_modules/.bin:$PATH"

ARG SERVER_URL
ENV VITE_SERVER_URL=$SERVER_URL
ARG RELATIVE_DASHBOARD_URL
ENV VITE_RELATIVE_DASHBOARD_URL=$RELATIVE_DASHBOARD_URL

COPY . .
RUN npm run build

# Production stage
FROM node:24-alpine AS production

RUN adduser -D -s /bin/sh -u 1001 appuser

WORKDIR /app

COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/package-lock.json ./package-lock.json

RUN npm ci --omit=dev && chown -R appuser:appuser /app

USER appuser

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1

CMD ["node", "build"]
