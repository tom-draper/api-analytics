import os

from dotenv import load_dotenv
from flask import Flask

from api_analytics.flask import add_middleware


load_dotenv()
api_key = os.environ.get("API_KEY")

app = Flask(__name__)
add_middleware(app, api_key)


@app.get("/")
def root():
    return {"message": "Hello World!"}


if __name__ == "__main__":
    app.run()
