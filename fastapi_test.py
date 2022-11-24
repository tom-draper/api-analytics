import os

from dotenv import load_dotenv
from fastapi import FastAPI

from analytics.fastapi import Analytics

load_dotenv()
api_key = os.environ.get("API_KEY")

app = FastAPI()
app.add_middleware(Analytics, api_key)


@app.get("/")
async def test():
    return "Test 1"


@app.get("/test/")
async def test():
    return "Test 2"
