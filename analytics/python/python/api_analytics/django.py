import threading
from time import time
from typing import Callable

from api_analytics.core import log_request
from django import conf
from django.core.handlers.wsgi import WSGIRequest
from django.http.response import HttpResponse


class Analytics:
    def __init__(self, get_response: Callable[[WSGIRequest], HttpResponse]) -> None:
        self.get_response = get_response
        self.api_key = getattr(conf.settings, "ANALYTICS_API_KEY", None)

    def __call__(self, request: WSGIRequest) -> HttpResponse:
        start = time()
        response = self.get_response(request)

        json = {
            'api_key': self.api_key,
            'hostname': request.get_host(),
            'path': request.path,
            'user_agent': request.headers['user-agent'],
            'method': request.method,
            'status': response.status_code,
            'framework': 9,
            'response_time': int((time() - start) * 1000),
        }

        threading.Thread(target=log_request, args=(json,)).start()
        return response