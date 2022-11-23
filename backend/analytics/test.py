import os

from analytics import Analytics
from dotenv import load_dotenv
from fastapi import FastAPI

load_dotenv()

api_key = os.environ.get("API_KEY")

app = FastAPI()
app.add_middleware(Analytics, api_key=api_key)


@app.get("/")
async def test():
    return "Test 2"


@app.get("/test/")
async def test():
    return "Test 2"
