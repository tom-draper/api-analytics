import os

import falcon
import uvicorn
from api_analytics.falcon import Analytics
from dotenv import load_dotenv

load_dotenv()

api_key = os.environ.get("API_KEY")


class Root:
    def on_get(self, req, resp):
        resp.media = {"message": "Hello World!"}


app = falcon.asgi.App(middleware=[Analytics(api_key)])
app.add_route("/", Root())

if __name__ == "__main__":
    uvicorn.run("app:app", reload=True)
