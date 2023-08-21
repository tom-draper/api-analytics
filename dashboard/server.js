const path = require("path");
const express = require("express");
const app = require("./public/App.js");

const server = express();

server.use(express.static(path.join(__dirname, "public")));

server.get("*", function(req, res) {
  const { html } = app.render({ url: req.url });

  res.write(`
    <!DOCTYPE html>
    <head>
      <link rel='stylesheet' href='/global.css'>
      <link rel='stylesheet' href='/bundle.css'>
      <link rel="icon" type="image/x-icon" href="/img/favicon.ico">
      <meta name="google-site-verification" content="vW5m_ij7wc85Td5Syr9N81a1efg_TTwcN8kOiXBc5c0" />
      <title>API Analytics</title>
      <script src="https://cdn.plot.ly/plotly-latest.min.js" type="text/javascript"></script>
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
    </head>

    <body>
      <div id="app">${html}</div>
      <script src="/bundle.js"></script>
    </body>
  `);

  res.end();
});

const port = 3000;
server.listen(port, () => console.log(`Listening on http://localhost:${port}`));
