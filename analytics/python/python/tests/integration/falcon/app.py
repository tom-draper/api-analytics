import os
import sys

# Import ../api_analytics rather than api_analytics pip package
sys.path.insert(0, os.path.abspath("../../../"))
import logging

import falcon
import uvicorn
from api_analytics.falcon import Analytics, Config
from dotenv import load_dotenv

logging.basicConfig(level=logging.DEBUG)

load_dotenv()

api_key = os.environ.get("API_KEY")


class Root:
    def on_get(self, req, resp):
        resp.media = {"message": "Hello World!"}


config = Config(
    get_ip_address=lambda req: req.get_header("x-forwarded-for", req.remote_addr)
)

app = falcon.asgi.App(middleware=[Analytics(api_key, config)])
app.add_route("/", Root())

if __name__ == "__main__":
    uvicorn.run("app:app", reload=True)
