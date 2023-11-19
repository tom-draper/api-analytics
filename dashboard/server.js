import { join } from 'path';
import express from 'express';
import app from './public/App.js';
import { fileURLToPath } from 'url';
import { dirname } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const port = 3000;
const server = express();

server.use(express.static(join(__dirname, 'public')));

server.get('*', function (req, res) {
    const { html } = app.render({ url: req.url });

    res.write(`
    <!DOCTYPE html>
    <head>
      <title>API Analytics</title>
      <meta name="description" content="Lightweight monitoring and analytics for API frameworks."/>
      <meta name="keywords" content="API Analytics, FastAPI, Express, Analytics, Dashboard, API, server"/>
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <meta name="google-site-verification" content="vW5m_ij7wc85Td5Syr9N81a1efg_TTwcN8kOiXBc5c0" />
      <link rel='stylesheet' href='/global.css'>
      <link rel='stylesheet' href='/bundle.css'>
      <link rel="icon" type="image/x-icon" href="/img/favicon.ico">
      <script src="https://cdn.plot.ly/plotly-latest.min.js" type="text/javascript"></script>
    </head>

    <body>
      <div id="app">${html}</div>
      <script src="/bundle.js"></script>
    </body>
  `);

    res.end();
});

server.listen(port, () => console.log(`Listening on http://localhost:${port}`));
