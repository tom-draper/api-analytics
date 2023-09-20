from datetime import datetime
from time import time

from api_analytics.core import log_request
from flask import Flask, request


def add_middleware(app: Flask, api_key: str):
    start  = 0.

    @app.before_request
    def prepare():
        global start
        start = time()


    @app.after_request
    def on_finish(response):
        global start
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

        log_request(api_key, request_data, 'Flask')
        return response