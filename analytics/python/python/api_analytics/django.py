from datetime import datetime
from time import time
from typing import Callable

from api_analytics.core import log_request
from django import conf
from django.core.handlers.wsgi import WSGIRequest
from django.http.response import HttpResponse


class Analytics:
    def __init__(self, get_response: Callable[[WSGIRequest], HttpResponse]):
        self.get_response = get_response
        self.api_key = getattr(conf.settings, "ANALYTICS_API_KEY", None)

    def __call__(self, request: WSGIRequest) -> HttpResponse:
        start = time()
        response = self.get_response(request)

        data = {
            'api_key': self.api_key,
            'hostname': request.get_host(),
            'ip_address': request.META.get('REMOTE_ADDR'),
            'path': request.path,
            'user_agent': request.headers['user-agent'],
            'method': request.method,
            'status': response.status_code,
            'framework': 'Django',
            'response_time': int((time() - start) * 1000),
            'created_at': datetime.now().isoformat()
        }

        log_request(data)
        return response
