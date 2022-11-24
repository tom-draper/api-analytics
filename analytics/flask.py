import threading
from time import time
from typing import Callable

from flask import Flask, Request, Response
from flask_http_middleware import BaseHTTPMiddleware, MiddlewareManager

from analytics.core import log_request


def add_middleware(app: Flask, api_key: str) -> None:
    app.wsgi_app = MiddlewareManager(app)
    app.wsgi_app.add_middleware(Analytics, api_key=api_key)


class Analytics(BaseHTTPMiddleware):
    def __init__(self, api_key: str):
        super().__init__()
        self.api_key = api_key

    def dispatch(self, request: Request, call_next: Callable):
        start = time()
        response: Response = call_next(request)
        elapsed = time() - start

        json = {
            'api_key': self.api_key,
            'hostname': request.host,
            'path': request.path,
            'user_agent': request.headers['user-agent'],
            'method': request.method,
            'status': int(response.status_code),
            'response_time': int(elapsed * 1000),
            'framework': 'Flask'
        }
        threading.Thread(target=log_request, args=(json,)).start()
        return response
