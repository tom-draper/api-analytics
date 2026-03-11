import os
import sys

# Import ../api_analytics rather than api_analyics pip package
sys.path.insert(0, os.path.abspath("../../../"))
import logging

import uvicorn
from api_analytics.starlette import Analytics, Config
from dotenv import load_dotenv
from starlette.applications import Starlette
from starlette.requests import Request
from starlette.responses import JSONResponse
from starlette.routing import Route

logging.basicConfig(level=logging.DEBUG)

load_dotenv()

api_key = os.environ.get("API_KEY")


async def root(request: Request):
    return JSONResponse({"message": "Hello World!"})


config = Config(
    get_ip_address=lambda request: request.headers.get(
        "X-Forwarded-For", request.client.host if request.client else None
    )
)

app = Starlette(routes=[Route("/", root)])
app.add_middleware(Analytics, api_key=api_key, config=config)

if __name__ == "__main__":
    uvicorn.run("app:app", reload=True)
