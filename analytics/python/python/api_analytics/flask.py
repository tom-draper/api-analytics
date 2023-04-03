from datetime import datetime
from time import time
from typing import Callable

from api_analytics.core import log_request
from flask import Flask, Request, Response
from flask_http_middleware import BaseHTTPMiddleware, MiddlewareManager


def add_middleware(app: Flask, api_key: str):
    app.wsgi_app = MiddlewareManager(app)
    app.wsgi_app.add_middleware(Analytics, api_key=api_key)


class Analytics(BaseHTTPMiddleware):
    def __init__(self, api_key: str):
        super().__init__()
        self.api_key = api_key

    def dispatch(self, request: Request, call_next: Callable[[Request], Response]):
        start = time()
        response = call_next(request)

        request_data = {
            'hostname': request.host,
            'ip_address': request.remote_addr,
            'path': request.path,
            'user_agent': request.headers['user-agent'],
            'method': request.method,
            'status': response.status_code,
            'response_time': int((time() - start) * 1000),
            'created_at': datetime.now().isoformat()
        }

        log_request(self.api_key, request_data, 'Flask')
        return response
