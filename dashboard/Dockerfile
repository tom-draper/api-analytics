# Étape de construction
FROM node:20-alpine AS builder

WORKDIR /app

# Copier les fichiers de dépendances
COPY package.json package-lock.json ./
RUN npm ci
ENV PATH="/app/node_modules/.bin:$PATH"

ARG SERVER_URL
ENV VITE_SERVER_URL=$SERVER_URL
ARG RELATIVE_DASHBOARD_URL
ENV VITE_RELATIVE_DASHBOARD_URL=$RELATIVE_DASHBOARD_URL

# Copier le reste des fichiers du projet
COPY . .

# Construire l'application
RUN npm run build

# Étape de production
FROM node:20-alpine AS production

# needed for healthcheck
RUN apk add --no-cache curl

WORKDIR /app

# Copier les fichiers nécessaires depuis l'étape de construction
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/package-lock.json ./package-lock.json

# Installer uniquement les dépendances de production
RUN npm ci --omit=dev

# Commande pour démarrer le serveur
CMD ["node", "build"]
