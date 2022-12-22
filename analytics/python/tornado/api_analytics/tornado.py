import threading
from time import time

from api_analytics.core import log_request
from tornado.web import RequestHandler


class Analytics(RequestHandler):
    def __init__(self, app, res, api_key):
        super().__init__(app, res)
        self.api_key = api_key
        self.start = None

    async def prepare(self):
        self.start = time()

    def on_finish(self):
        data = {
            'api_key': self.api_key,
            'hostname': self.request.host,
            'path': self.request.path,
            'user_agent': self.request.headers['user-agent'],
            'method': self.request.method,
            'status': self.get_status(),
            'framework': 'Tornado',
            'response_time': int((time() - self.start) * 1000),
        }

        threading.Thread(target=log_request, args=(data,)).start()
        self.start = None
