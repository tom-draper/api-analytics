import os

from api_analytics.tornado import Analytics
from dotenv import load_dotenv
from tornado.ioloop import IOLoop
from tornado.web import Application

load_dotenv()


class MainHandler(Analytics):
    def __init__(self, app, res):
        api_key = os.environ.get("API_KEY")
        super().__init__(app, res, api_key)

    def get(self):
        self.write({'message': 'Hello World!'})


def make_app():
    return Application([
        (r"/", MainHandler),
    ])


if __name__ == "__main__":
    app = make_app()
    app.listen(8000)
    IOLoop.instance().start()