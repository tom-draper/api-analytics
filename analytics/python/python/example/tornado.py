import asyncio
from tornado.web import Application

from api_analytics.tornado import Analytics


class MainHandler(Analytics):
    def __init__(self, app, res):
        super().__init__(app, res, "hello")

    def get(self):
        self.write({'message': 'Hello World!'})


def make_app():
    return Application([
        (r"/", MainHandler),
    ])


async def main():
    app = make_app()
    app.listen(8080)
    await asyncio.Event().wait()

if __name__ == "__main__":
    asyncio.run(main())
