import sys
import os
sys.path.insert(0, os.path.abspath('../'))

from dotenv import load_dotenv
from fastapi import FastAPI

from api_analytics.fastapi import Analytics

load_dotenv()
api_key = os.environ.get("API_KEY")

app = FastAPI()
app.add_middleware(Analytics, api_key=api_key)


@app.get("/")
async def test():
    return "Test 1"


@app.get("/test/")
async def test():
    return "Test 2"

if __name__ == '__main__':
    app.run()