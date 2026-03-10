import os

import uvicorn
from api_analytics.starlette import Analytics
from dotenv import load_dotenv
from starlette.applications import Starlette
from starlette.requests import Request
from starlette.responses import JSONResponse
from starlette.routing import Route

load_dotenv()

api_key = os.environ.get("API_KEY")


async def root(request: Request):
    return JSONResponse({"message": "Hello World!"})


app = Starlette(routes=[Route("/", root)])
app.add_middleware(Analytics, api_key=api_key)

if __name__ == "__main__":
    uvicorn.run("app:app", reload=True)
