# Self-Hosting API Analytics

API Analytics can be easily self-hosted giving you full control over your stored logged request data.

Self-hosting requires:

- an environment that can run Docker Compose and is publically addressable, such as a VPS; and
- a domain name that can be pointed to your server's IP address.

By default, the `docker-compose.yml` file is set up to generate a free SSL certificate for your domain using Certbot and Let's Encrypt.

You may need to adjust this configuration to work with your environment.

**Self-hosting is still undergoing testing, development and further improvements to make it as easy as possible to deploy. It is currently recommended that you avoid self-hosting for production use.**

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

#### 2. Edit the `.env` file

Enter:
- your `DOMAIN_NAME` e.g. example.com 
- a `POSTGRES_PASSWORD` for the database

#### 3. Obtain an SSL certificate using Certbot

Start the `nginx` server.

```bash
docker compose up nginx -d
```

Generate the SSL certificate, replacing `your-domain.com` with your actual domain and `your-email@example.com` with your email address.

```bash
docker compose run --rm certbot certonly --webroot --webroot-path=/var/www/certbot -d your-domain.com -d www.your-domain.com --agree-tos --email your-email@example.com --no-eff-email
```

Stop the `nginx` server once complete.

```bash
docker compose down nginx
```

Replace the temporary `nginx-certbot.conf.template` with the fully SSL-compatible `nginx.conf.template` in `docker-compose.yaml` under the `nginx` configuration. Comment out the appropriate lines to match the following:

```yaml
# - ./nginx/nginx-certbot.conf.template:/etc/nginx/conf.d/nginx.conf.template
- ./nginx/nginx.conf.template:/etc/nginx/conf.d/nginx.conf.template
```

#### 4. Start the services

```bash
docker compose up -d
```

#### Testing

##### Internal

Check if all six docker services are running.

```bash
docker ps
```

Quickly check if services are working by attempting to generate a new API key.

```bash
curl -X GET http://localhost:3000/api/generate
```

Confirm services are working internally by running the `tests/test-internal.sh` bash script.

```bash
chmod +x tests/test-internal.sh
./tests/test-internal.sh
```

##### Nginx

Confirm Nginx is running and able to direct to the internal services.

```bash
curl -kL -X GET http://localhost/api/generate
```

```bash
curl -k -X GET https://localhost/api/generate
```

##### External

Outside of the hosting environment, confirm that services are publically accessible with an API key generation attempt.

```bash
curl -X GET http://<ip-address>:3000/api/generate
```

Confirm your domain is set up and that Nginx is working correctly.

```bash
curl -X GET https://your-domain.com/api/generate
```

Finally confirm the dashboard can communicate with your server by attempting to generate an API key at: `https://www.apianalytics.dev/generate?source=https://your-domain.com`

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

### Usage

#### Logging Requests

Once your backend services are up and running, you can log requests to your server by specifying the server URL within the API Analytics middleware config.

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

You can confirm requests are being logged by checking the logs.

```bash
docker exec -it logger tail requests.log
```

#### Dashboard

You can use the dashboard by specifying the URL of your server as a `source` parameter when using `apianalytics.dev`, or you can access the raw data directly by making a GET request to your API data endpoint.

You can access your dashboard at: `https://www.apianalytics.dev/dashboard?source=https://www.your-domain.com`

You can access your raw data by sending a GET request to `https://www.your-domain.com/api/data`, with your API key set as `X-AUTH-TOKEN` in the headers.

## Frontend Hosting

Once up and running, self-hosted backend can be fully utilised and managed through `apianalytics.dev`, and this will ensure you always have the latest updates and improvements to the dashboard. The frontend can also be self-hosted by setting storing the URL of your backend service in the environment variable `SERVER_URL`, or manually changing the `SERVER_URL` held in `src/lib/consts.ts`. The frontend can then be deploying with your favourite hosting provider.
