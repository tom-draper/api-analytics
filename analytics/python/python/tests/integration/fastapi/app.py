import os
import sys

# Import ../api_analytics rather than api_analyics pip package
sys.path.insert(0, os.path.abspath("../"))
import logging

import uvicorn
from api_analytics.fastapi import Analytics, Config
from dotenv import load_dotenv
from fastapi import FastAPI

logging.basicConfig(level=logging.DEBUG)

load_dotenv()

api_key = os.environ.get("API_KEY")

app = FastAPI()
config = Config()
config.get_ip_address = lambda request: request.headers.get(
    "X-Forwarded-For", request.client.host
)
app.add_middleware(Analytics, api_key=api_key, config=config)


@app.get("/")
async def root():
    return {"message": "Hello World!"}


if __name__ == "__main__":
    uvicorn.run("app:app", reload=True)
