from dataclasses import dataclass
from datetime import datetime
from time import time
from typing import Callable

from api_analytics.core import log_request
from tornado.httputil import HTTPServerRequest
from tornado.web import Application, RequestHandler


@dataclass
class Config:
    get_path: Callable[[HTTPServerRequest], str] | None = None
    get_ip_address: Callable[[HTTPServerRequest], str] | None = None
    get_hostname: Callable[[HTTPServerRequest], str] | None = None
    get_user_agent: Callable[[HTTPServerRequest], str] | None = None
    get_user_id: Callable[[HTTPServerRequest], str] | None = None


class Analytics(RequestHandler):
    def __init__(self, app: Application, res: HTTPServerRequest, api_key: str, config: Config = Config()):
        super().__init__(app, res)
        self.api_key = api_key
        self.config = config
        self.start = time()

    def prepare(self):
        self.start = time()

    def on_finish(self):
        request_data = {
            'hostname': self._get_hostname(),
            'ip_address': self._get_ip_address(),
            'path': self._get_path(),
            'user_agent': self.get_user_agent(),
            'method': self.request.method,
            'status': self.get_status(),
            'response_time': int((time() - self.start) * 1000),
            'created_at': datetime.now().isoformat()
        }

        log_request(self.api_key, request_data, 'Tornado')
        self.start = None

    def _get_path(self) -> str | None:
        if self.config.get_path:
            return self.config.get_path(self.request)
        return self.request.path

    def _get_ip_address(self) -> str | None:
        if self.config.get_ip_address:
            return self.config.get_ip_address(self.request)
        return self.request.remote_ip

    def _get_hostname(self) -> str | None:
        if self.config.get_hostname:
            return self.config.get_hostname(self.request)
        return self.request.host

    def _get_user_id(self) -> str | None:
        if self.config.get_user_id:
            return self.config.get_user_id(self.request)
        return None

    def _get_user_agent(self) -> str | None:
        if self.config.get_user_agent:
            return self.config.get_user_agent(self.request)
        elif 'user-agent' in self.request.headers:
            return self.request.headers['user-agent']
        elif 'User-Agent' in self.request.headers:
            return self.request.headers['User-Agent']
        return None
