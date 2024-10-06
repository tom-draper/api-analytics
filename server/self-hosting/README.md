# Self-Hosting API Analytics

API Analytics can be easily self-hosted giving you full control over your stored logged request data. Self-hosting requires an environment that can run Docker Compose and is publically addressable over HTTPS such as a VPS. You will also need a domain name that can be pointed to your server's IP address. By default, the `docker-compose.yml` file is set up to generate a free SSL certificate for your domain using Certbot and Let's Encrypt.

## Backend Hosting

### Getting Started

```bash
docker compose up -d
```

```bash
docker compose stop
```

### Usage

Once your backend service is live and working, it can be consumed by specifying the URL as a `source` parameter when using `apianalytics.dev`, or by making direct requests to your own API to access your raw data.

You can access your dashboard at: `https://www.apianalytics.dev/dashboard?source=<https://www.yourdomain.com>`

You can access your raw data at: `https://www.yourdomain.com/api/data`

## Frontend Hosting

Once up and running, self-hosted backend can be fully utilised and managed through `apianalytics.dev`, and this will ensure you always have the latest updates and improvements to the dashboard. The frontend can also be self-hosted by setting storing the URL of your backend service in the environment variable `SERVER_URL`, or manually changing the `SERVER_URL` held in `src/lib/consts.ts`. The frontend can then be deploying with your favourite hosting provider.
