import os
import sys

# Import ../api_analytics rather than api_analytics pip package
sys.path.insert(0, os.path.abspath("../../../"))
import logging

from api_analytics.sanic import add_middleware, Config
from dotenv import load_dotenv
from sanic import Sanic
from sanic.response import json

logging.basicConfig(level=logging.DEBUG)

load_dotenv()

api_key = os.environ.get("API_KEY")

app = Sanic("app")

config = Config(
    get_ip_address=lambda request: request.headers.get(
        "X-Forwarded-For", request.ip
    )
)
add_middleware(app, api_key, config)


@app.get("/")
async def root(request):
    return json({"message": "Hello World!"})


if __name__ == "__main__":
    app.run()
