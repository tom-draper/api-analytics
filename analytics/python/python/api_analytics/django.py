from dataclasses import dataclass
from datetime import datetime
from time import time
from typing import Callable

from api_analytics.core import log_request
from django import conf
from django.core.handlers.wsgi import WSGIRequest
from django.http.response import HttpResponse


@dataclass
class Config:
    get_path: Callable[[WSGIRequest], str] | None = None
    get_ip_address: Callable[[WSGIRequest], str] | None = None
    get_hostname: Callable[[WSGIRequest], str] | None = None
    get_user_agent: Callable[[WSGIRequest], str] | None = None
    get_user_id: Callable[[WSGIRequest], str] | None = None


class Analytics:
    def __init__(self, get_response: Callable[[WSGIRequest], HttpResponse]):
        self.get_response = get_response
        self.api_key = getattr(conf.settings, "ANALYTICS_API_KEY", None)
        self.config = getattr(conf.settings, "ANALYTICS_CONFIG", Config())

    def __call__(self, request: WSGIRequest) -> HttpResponse:
        start = time()
        response = self.get_response(request)

        request_data = {
            'hostname': self._get_hostname(request),
            'ip_address': self._get_ip_address(request),
            'path': self._get_path(request),
            'user_agent': self._get_user_agent(request),
            'method': request.method,
            'status': response.status_code,
            'response_time': int((time() - start) * 1000),
            'created_at': datetime.now().isoformat()
        }

        log_request(self.api_key, request_data, 'Django')
        return response

    def _get_path(self, request: WSGIRequest) -> str | None:
        if self.config.get_path:
            return self.config.get_path(request)
        return request.path
    
    def _get_ip_address(self, request: WSGIRequest) -> str | None:
        if self.config.get_ip_address:
            return self.config.get_ip_address(request)
        return request.META.get('REMOTE_ADDR')
    
    def _get_hostname(self, request: WSGIRequest) -> str | None:
        if self.config.get_hostname:
            return self.config.get_hostname(request)
        return request.get_host()
    
    def _get_user_id(self, request: WSGIRequest) -> str | None:
        if self.config.get_user_id:
            return self.config.get_user_id(request)
        return None
    
    def _get_user_agent(self, request: WSGIRequest) -> str | None:
        if self.config.get_user_agent:
            return self.config.get_user_agent(request)
        elif 'user-agent' in request.headers:
            return request.headers['user-agent']
        elif 'User-Agent' in request.headers:
            return request.headers['User-Agent']
        return None