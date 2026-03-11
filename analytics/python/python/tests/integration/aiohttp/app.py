import os
import sys

# Import ../api_analytics rather than api_analyics pip package
sys.path.insert(0, os.path.abspath("../../../"))
import logging

from aiohttp import web
from api_analytics.aiohttp import add_middleware, Config
from dotenv import load_dotenv

logging.basicConfig(level=logging.DEBUG)

load_dotenv()

api_key = os.environ.get("API_KEY")


async def root(request: web.Request) -> web.Response:
    return web.json_response({"message": "Hello World!"})


config = Config(
    get_ip_address=lambda request: request.headers.get(
        "X-Forwarded-For", request.remote
    )
)

app = web.Application()
add_middleware(app, api_key, config)
app.router.add_get("/", root)

if __name__ == "__main__":
    web.run_app(app)
