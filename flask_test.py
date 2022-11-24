import os

from dotenv import load_dotenv
from flask import Flask

from analytics.flask import add_middleware


load_dotenv()
api_key = os.environ.get("API_KEY")

app = Flask(__name__)
add_middleware(app, api_key)


@app.get("/health")
def health():
    return {"message": "I'm healthy"}


if __name__ == "__main__":
    app.run()
