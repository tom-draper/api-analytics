# Go-Gin Vercel Starter

Starter untuk menjalankan web framework [Gin](https://github.com/gin-gonic/gin) pada serverless [Vercel](https://vercel.com/).

## Demo

- <https://go-gin-vercel.vercel.app/api/hi>
- <https://go-gin-vercel.vercel.app/api/ping>

## Deploy your own

Deploy the example using [Vercel](https://vercel.com):

[![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/git/external?repository-url=https://github.com/zakiego/go-gin-vercel&project-name=go-gin-vercel&repository-name=go-gin-vercel)

## How to use

1. Clone repository

```bash
git clone https://github.com/zakiego/go-gin-vercel

cd go-gin-vercel
```

2. Install [Vercel CLI](https://vercel.com/docs/clihttps://vercel.com/docs/cli)

```bash
yarn global add vercel

or

npm i -g vercel
```

3. Install package.json

```bash
yarn install
```

4. Menjalankan di local

```bash
yarn dev
```

5. Deploy ke vercel

```bash
yarn deploy
```

6. Deploy ke Vercel versi [production](https://vercel.com/docs/cli#introduction/unique-options/prod)

```bash
yarn deploy:prod
```

## Reference

- <https://vercel.com/docs/project-configuration>
- <https://vercel.com/docs/runtimes#official-runtimes/go>
- <https://vercel.com/docs/cli>
- <https://github.com/kirito41dd/vercel-faas>
