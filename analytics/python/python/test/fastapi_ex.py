import os
import sys
# Import ../api_analytics rather than api_analyics pip package
sys.path.insert(0, os.path.abspath('../'))
from api_analytics.fastapi import Analytics

import uvicorn
from fastapi import FastAPI

from dotenv import load_dotenv

load_dotenv()


api_key = os.environ.get("API_KEY")

app = FastAPI()
app.add_middleware(Analytics, api_key=api_key)


@app.get("/")
async def root():
    return {'message': 'Hello World!'}

if __name__ == "__main__":
    uvicorn.run("fastapi_ex:app", reload=True)