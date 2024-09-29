# Self-Hosting API Analytics

API Analytics can be self-hosted giving you full control over your stored logged request data. Self-hosting requires an environment that can run docker-compose and is publically addressable over HTTPS such as a VPS.

## Backend Hosting

### Getting Started

Update `nginx/nginx.conf` to replace `example.com` with your URL.



### Usage

Once your backend service is live and working, it can be consumed by specifying the URL as a source when using `apianalytics.dev`, or by making requests to your API to access your raw data.

You can access your dashboard at: `https://www.apianalytics.dev/dashboard?source=<https://www.example.com>`

You can access your raw data at: `https://www.example.com/api/data`

## Frontend Hosting

A self-hosted backend can be fully utilised through `apianalytics.dev`, and this will ensure you always have the latest updates and improvements to the dashboard. The frontend can also be self-hosted by updating the `SERVER_URL` held in `src/lib/consts.ts` and deploying with your favourite hosting provider.