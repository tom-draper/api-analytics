# Self-Hosting API Analytics

API Analytics can be easily self-hosted giving you full control over your stored logged request data.

Self-hosting requires:

- an environment that can run Docker Compose and is publically addressable, such as a VPS; and
- a domain name that can be pointed to your server's IP address.

By default, the `docker-compose.yml` file is set up to generate a free SSL certificate for your domain using Certbot and Let's Encrypt.

**Self-hosting is still undergoing testing, development and further improvements to make it easy to deploy. It is currently recommended that you avoid self-hosting for production use.**

## Backend Hosting

### Getting Started

Clone the repo.

```bash
git clone github.com/tom-draper/api-analytics
```

Open the `self-hosting` directory.

```bash
cd api-analytics/server/self-hosting
```

Enter your `DOMAIN_NAME` as an environment variable within `.env`.

Start the services.

```bash
docker compose up -d
```

#### Testing

##### Internal

You can quickly confirm if services are working by generating a new API key.

```bash
curl -X GET http://localhost:3000/api/generate
```

Confirm services are working internally by running `test.sh` bash script.

```bash
chmod +x test.sh
./test.sh
```

##### External

Outside of the hosting environment, you can confirm that services are publically accessible with an API key generation attempt.

```bash
curl -X GET <IP-ADDRESS>:3000/api/generate
```

Confirm your domain is set up and that Nginx is working correctly.

```bash
curl -X GET <DOMAIN-NAME>/api/generate
```

#### Maintenance

Check the status of the running services with:

```bash
docker ps
```

If needed, you can stop all services with:

```bash
docker compose stop
```

Remove all containers and images with:

```bash
docker compose down --rmi all
```

### Usage

Once your backend service is live and working, it can be consumed by specifying the URL as a `source` parameter when using `apianalytics.dev`, or by making direct requests to your own API to access your raw data.

You can access your dashboard at: `https://www.apianalytics.dev/dashboard?source=<https://www.yourdomain.com>`

You can access your raw data at: `https://www.yourdomain.com/api/data`

## Frontend Hosting

Once up and running, self-hosted backend can be fully utilised and managed through `apianalytics.dev`, and this will ensure you always have the latest updates and improvements to the dashboard. The frontend can also be self-hosted by setting storing the URL of your backend service in the environment variable `SERVER_URL`, or manually changing the `SERVER_URL` held in `src/lib/consts.ts`. The frontend can then be deploying with your favourite hosting provider.
