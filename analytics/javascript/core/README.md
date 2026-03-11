# @api-analytics/core

Shared core package for the API Analytics JavaScript middleware packages. This package is not intended to be installed directly — it is a dependency of the framework-specific packages.

## Packages

| Framework | Package |
|-----------|---------|
| Express | [@api-analytics/express](https://www.npmjs.com/package/@api-analytics/express) |
| Fastify | [@api-analytics/fastify](https://www.npmjs.com/package/@api-analytics/fastify) |
| Koa | [@api-analytics/koa](https://www.npmjs.com/package/@api-analytics/koa) |
| Hono | [@api-analytics/hono](https://www.npmjs.com/package/@api-analytics/hono) |
| NestJS | [@api-analytics/nestjs](https://www.npmjs.com/package/@api-analytics/nestjs) |
| Elysia | [@api-analytics/elysia](https://www.npmjs.com/package/@api-analytics/elysia) |
| Bun | [@api-analytics/bun](https://www.npmjs.com/package/@api-analytics/bun) |
| Oak (Deno) | [@api-analytics/oak](https://jsr.io/@api-analytics/oak) |

## Exports

- `Config` — Configuration class with optional mapper functions to override default request data extraction
- `Mappers` — Default mapper implementations for Node.js-style request objects
- `Analytics` — Core analytics client that batches and posts request data to the server
- `getIPAddress` — Helper that resolves the client IP address respecting the configured privacy level

## Contributions

Contributions, issues and feature requests are welcome.

- Fork it (https://github.com/tom-draper/api-analytics)
- Create your feature branch (`git checkout -b my-new-feature`)
- Commit your changes (`git commit -am 'Add some feature'`)
- Push to the branch (`git push origin my-new-feature`)
- Create a new Pull Request

---

If you find value in my work consider supporting me.

Buy Me a Coffee: https://www.buymeacoffee.com/tomdraper<br>
PayPal: https://www.paypal.com/paypalme/tomdraper
