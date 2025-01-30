import os
import sys

# Import ../api_analytics rather than api_analyics pip package
sys.path.insert(0, os.path.abspath("../../../"))
from api_analytics.flask import add_middleware

from flask import Flask

from dotenv import load_dotenv
import logging

logging.basicConfig(level=logging.DEBUG)

load_dotenv()


api_key = os.environ.get("API_KEY")

app = Flask(__name__)
add_middleware(app, api_key)


@app.get("/")
def root():
    return {"message": "Hello World!"}


if __name__ == "__main__":
    app.run()
