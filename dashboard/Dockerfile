# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY package.json package-lock.json ./
RUN npm ci
ENV PATH="/app/node_modules/.bin:$PATH"

ARG SERVER_URL
ENV VITE_SERVER_URL=$SERVER_URL
ARG RELATIVE_DASHBOARD_URL
ENV VITE_RELATIVE_DASHBOARD_URL=$RELATIVE_DASHBOARD_URL

# Copy the rest of the project files
COPY . .

# Build the application
RUN npm run build

# Production stage
FROM node:20-alpine AS production

# needed for healthcheck
RUN apk add --no-cache curl

WORKDIR /app

# Copy built output from build stage
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/package-lock.json ./package-lock.json

# Install production dependencies only
RUN npm ci --omit=dev

# Start the server
CMD ["node", "build"]
