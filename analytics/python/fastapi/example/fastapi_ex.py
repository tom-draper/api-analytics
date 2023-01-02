import os

import uvicorn
from api_analytics.fastapi import Analytics
from dotenv import load_dotenv
from fastapi import FastAPI

load_dotenv()


api_key = os.environ.get("API_KEY")

app = FastAPI()
app.add_middleware(Analytics, api_key=api_key)


@app.get("/")
async def root():
    return {'message': 'Hello World!'}

if __name__ == "__main__":
    uvicorn.run("fastapi_ex:app", reload=True)