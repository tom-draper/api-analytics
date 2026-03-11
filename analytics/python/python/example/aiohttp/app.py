import os

from aiohttp import web
from api_analytics.aiohttp import add_middleware
from dotenv import load_dotenv

load_dotenv()

api_key = os.environ.get("API_KEY")


async def root(request: web.Request) -> web.Response:
    return web.json_response({"message": "Hello World!"})


app = web.Application()
add_middleware(app, api_key)
app.router.add_get("/", root)

if __name__ == "__main__":
    web.run_app(app)
