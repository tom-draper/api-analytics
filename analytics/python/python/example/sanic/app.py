import os

from api_analytics.sanic import add_middleware
from dotenv import load_dotenv
from sanic import Sanic
from sanic.response import json

load_dotenv()

api_key = os.environ.get("API_KEY")

app = Sanic("app")
add_middleware(app, api_key)


@app.get("/")
async def root(request):
    return json({"message": "Hello World!"})


if __name__ == "__main__":
    app.run()
