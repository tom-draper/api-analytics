from dataclasses import dataclass
from datetime import datetime
from time import time
from typing import Callable

from api_analytics.core import log_request
from starlette.middleware.base import (BaseHTTPMiddleware,
                                       RequestResponseEndpoint)
from starlette.requests import Request
from starlette.responses import Response
from starlette.types import ASGIApp


@dataclass
class Config:
    get_path: Callable[[Request], str] | None = None
    get_ip_address: Callable[[Request], str] | None = None
    get_hostname: Callable[[Request], str] | None = None
    get_user_agent: Callable[[Request], str] | None = None
    get_user_id: Callable[[Request], str] | None = None


class Analytics(BaseHTTPMiddleware):
    def __init__(self, app: ASGIApp, api_key: str, config: Config = Config()):
        super().__init__(app)
        self.api_key = api_key
        self.config = config

    def _get_user_agent(self, request: Request) -> str | None:
        if self.config.get_user_agent:
            return self.config.get_user_agent(request)
        elif 'user-agent' in request.headers:
            return request.headers['user-agent']
        elif 'User-Agent' in request.headers:
            return request.headers['User-Agent']
        return None

    def _get_path(self, request: Request) -> str | None:
        if self.config.get_path:
            return self.config.get_path(request)
        return request.url.path

    def _get_ip_address(self, request: Request) -> str | None:
        if self.config.get_ip_address:
            return self.config.get_ip_address(request)
        return request.client.host

    def _get_hostname(self, request: Request) -> str | None:
        if self.config.get_hostname:
            return self.config.get_hostname(request)
        return request.url.hostname

    def _get_user_id(self, request: Request) -> str | None:
        if self.config.get_user_id:
            return self.config.get_user_id(request)
        return None

    async def dispatch(self, request: Request, call_next: RequestResponseEndpoint) -> Response:
        start = time()
        response = await call_next(request)

        request_data = {
            'hostname': self._get_hostname(request),
            'ip_address': self._get_ip_address(request),
            'path': self._get_path(request),
            'user_agent': self._get_user_agent(request),
            'method': request.method,
            'status': response.status_code,
            'response_time': int((time() - start) * 1000),
            'created_at': datetime.now().isoformat(),
        }

        log_request(self.api_key, request_data, 'FastAPI')
        return response
