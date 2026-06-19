# Self-Hosting API Analytics

API Analytics can be easily self-hosted, allowing for full control over your logged request data.

Requirements:

- a publically addressable environment that can run Docker Compose, such as a VPS; and
- a domain name that is pointing to your server's IP address.

By default, the `docker-compose.yml` file is set up to generate a free SSL certificate for your domain using Certbot and Let's Encrypt.

You may need to adjust this configuration to work with your environment.

**Self-hosting is still being refined to make deployment as smooth as possible. Please test thoroughly before relying on it in production.**

## Backend Hosting

### Getting Started

#### 1. Clone the repo

```bash
git clone github.com/tom-draper/api-analytics
```

Open the `self-hosting` directory.

```bash
cd api-analytics/server/self-hosting
```

#### 2. Create a `.env` file

Create a new `.env` file, using the provided `.env.example` as a template.

Enter:
- your `DOMAIN_NAME` e.g. example.com
- a `POSTGRES_PASSWORD` for the database

#### 3. Obtain an SSL certificate using Certbot

Start the `certbot` and `nginx` services.

```bash
docker compose up certbot nginx -d
```

Generate the SSL certificate, replacing `your-domain.com` with your actual domain and `your-email@example.com` with your email address.

```bash
docker exec -it certbot certbot certonly --webroot -w /var/www/certbot -d your-domain.com -d www.your-domain.com --agree-tos --email your-email@example.com --no-eff-email
```

Stop the services once complete.

```bash
docker compose down
```

Within `docker-compose.yaml`, replace the temporary `nginx-certbot.conf.template` with the fully SSL-compatible `nginx.conf.template` under the `nginx` configuration. Comment the appropriate lines to match the following:

```yaml
# - ./nginx/nginx-certbot.conf.template:/etc/nginx/conf.d/nginx.conf.template
- ./nginx/nginx.conf.template:/etc/nginx/conf.d/nginx.conf.template
```

#### 4. Start the services

```bash
docker compose up -d
```

### Testing

#### Internal

Confirm all six Docker services are running internally.

```bash
docker ps
```

From the server, quickly check if internal services are working by attempting to generate a new API key.

```bash
curl -X GET http://localhost:3000/api/generate
```

For a more comprehensive test, confirm services are working internally by running the `tests/test-internal.sh` bash script.

```bash
chmod +x tests/test-internal.sh
./tests/test-internal.sh
```

#### Nginx

From the server, confirm Nginx is running and redirecting to the internal services correctly.

```bash
curl -kL -X GET http://localhost/api/generate
```

```bash
curl -k -X GET https://localhost/api/generate
```

For a more comprehensive test, confirm the Nginx service is working internally by running the `tests/test-nginx.sh` and  `tests/test-nginx-ssl.sh` bash scripts.

```bash
chmod +x tests/test-nginx.sh
./tests/test-nginx.sh

chmod +x tests/test-nginx-ssl.sh
./tests/test-nginx-ssl.sh
```

#### External

Outside of the hosting environment, confirm that services are publically accessible with an API key generation attempt.

```bash
curl -X GET http://<ip-address>:3000/api/generate
```

Confirm your domain is set up and that Nginx is redirecting correctly.

```bash
curl -X GET https://your-domain.com/api/generate
```

For a more comprehensive test, confirm the services are working externally by running the `tests/test-external.sh` bash script, providing your domain name.

```bash
chmod +x tests/test-external.sh
./tests/test-external.sh your-domain.com
```

Finally, confirm the dashboard can communicate with your server by attempting to generate an API key at: `https://www.apianalytics.dev/generate?source=https://your-domain.com`

You can check:
- Nginx logs with `docker logs nginx`
- API logs with `docker exec -it api tail api.log`

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

##### Database

The `database/schema.sql` schema is used to initialise the postgres database once the container is first built.

You can run custom SQL commands with:

```bash
docker exec -it db psql -U postgres -d analytics -c "YOUR SQL COMMAND;"
```

##### Updates

Updating the backend with the latest improvements is straight-forward, but will come with some downtime.

```bash
docker compose down
git pull origin main
docker compose up -d
```

##### Locations

Optional IP-to-location mappings are provided by the GeoLite2 Country database maintained by MaxMind. Create a free account at `https://www.maxmind.com/en/home`, and download and copy the `GeoLite2-Country.mmdb` file into the `server/logger` folder.

### Usage

#### Logging Requests

Once your backend services are running and tested, you can log requests to your server by specifying the server URL within the API Analytics middleware config.

```py
import uvicorn
from api_analytics.fastapi import Analytics, Config
from fastapi import FastAPI

app = FastAPI()
config = Config(server_url='https://your-domain.com')
app.add_middleware(Analytics, api_key=<api-key>, config=config)

@app.get("/")
async def root():
    return {"message": "Hello World!"}

if __name__ == "__main__":
    uvicorn.run("app:app", reload=True)
```

When debugging, checking the server logs is usually the best place to start.

```bash
docker logs nginx

docker exec -it logger tail requests.log

docker exec -it api tail api.log
```

#### Dashboard

The easiest way to view your self-hosted data is the hosted dashboard at `apianalytics.dev` — just point it at your backend with the `source` URL parameter, no deployment required:

```
https://www.apianalytics.dev/dashboard?source=https://your-domain.com
```

The dashboard fetches and renders your data entirely in your browser, directly from your backend — your API key and request data never pass through the hosted service. The `source` parameter is carried over when you sign in, so it applies to your dashboard, monitor and explorer views.

Alternatively, access your raw data directly with a GET request to `https://your-domain.com/api/data`, with your API key set as the `X-AUTH-TOKEN` header.

## Frontend Hosting

Most self-hosters can simply use the hosted dashboard with the `source` parameter described above — it always has the latest improvements and requires nothing to deploy.

If you'd rather host the dashboard yourself, point it at your backend with the `SERVER_URL` environment variable. Copy `dashboard/.env.example` to `dashboard/.env` and set:

```
SERVER_URL=https://your-domain.com
```

Then build and deploy with your preferred hosting provider:

```bash
cd api-analytics/dashboard
pnpm install
pnpm build
```

This uses `adapter-auto`, which targets common hosting platforms automatically.

#### Docker

To run the dashboard as a standalone container, build with the `node` adapter using the provided `Dockerfile`, passing your backend URL as a build argument:

```bash
cd api-analytics/dashboard
docker build --build-arg SERVER_URL=https://your-domain.com -t api-analytics-dashboard .
docker run -p 3000:3000 api-analytics-dashboard
```

The dashboard will be available at `http://localhost:3000`.

Alternatively, bring it up alongside the backend with the `docker-compose.yml` file. The dashboard is an opt-in service (it isn't started by the default `docker compose up`), enabled with the `dashboard` profile. It builds using your `DOMAIN_NAME` from the `.env` file as the backend URL and is served on port `5173`:

```bash
docker compose --profile dashboard up -d --build
```

`SERVER_URL` is baked into the build, so rebuild if it changes. The `source` URL parameter still works on a self-hosted dashboard and takes precedence, letting you override the backend per-visit.

## Contributions

Feel free to customise this project to your preference. Any feedback or improvements that can still generalise to most deployment environments is much appreciated.


## Alternative with Traefik

### Development environment example

```bash
cd api-analytics/server/self-hosting
ln -s .env.dev .env
docker compose -f docker-compose.traefik-dev-example.yml up -d
```

* Traefik's dashboard page is served at http://localhost:8080
* Dev Dashbaord is served at http://localhost:5173
* Built Dashbaord is served at http://localhost/build
* API is served at http://localhost/api (GET). `curl -X GET http://localhost/api/health`
* Logger is served at http://localhost/api (POST) `curl -X POST http://localhost/api/requests`


### Production environment example

```bash
cd api-analytics/server/self-hosting
ln -s .env.prod .env
# IMPORTANT : set <DOMAIN_NAME> variable with your own domain into .env.prod
#             docker-compose.traefik-prod-example.yml have to be served onto <DOMAIN_NAME> server
docker compose -f docker-compose.traefik-prod-example.yml up -d
```

* https certificates are auto generated through letsencrypt ACME 👍️
* Dashbaord is served at https://example.com/api-analytics
* API is served at https://example.com/analytics-backend/api (GET). `curl -X GET https://example.com/analytics-backend/api/health`
* Logger is served at https://example.com/analytics-backend/api (POST) `curl -X POST https://example.com/analytics-backend/api/requests`