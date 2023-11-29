from dataclasses import dataclass
from datetime import datetime
from time import time
from typing import Callable

from api_analytics.core import log_request
from flask import Flask, Request, Response, request


@dataclass
class Config:
    get_path: Callable[[Request], str] | None = None
    get_ip_address: Callable[[Request], str] | None = None
    get_hostname: Callable[[Request], str] | None = None
    get_user_agent: Callable[[Request], str] | None = None
    get_user_id: Callable[[Request], str] | None = None


def add_middleware(app: Flask, api_key: str, config: Config = Config()):
    start = 0.

    @app.before_request
    def prepare():
        nonlocal start
        start = time()

    @app.after_request
    def on_finish(response: Response) -> Response:
        nonlocal start
        request_data = {
            'hostname': _get_hostname(request, config),
            'ip_address': _get_ip_address(request, config),
            'path': _get_path(request, config),
            'user_agent': _get_user_agent(request),
            'method': request.method,
            'status': response.status_code,
            'response_time': int((time() - start) * 1000),
            'created_at': datetime.now().isoformat()
        }

        log_request(api_key, request_data, 'Flask')
        return response

def _get_path(request: Request, config: Config) -> str | None:
    if config.get_path:
        return config.get_path(request)
    return request.path

def _get_ip_address(request: Request, config: Config) -> str | None:
    if config.get_ip_address:
        return config.get_ip_address(request)
    return request.remote_addr

def _get_hostname(request: Request, config: Config) -> str | None:
    if config.get_hostname:
        return config.get_hostname(request)
    return request.host

def _get_user_id(request: Request, config: Config) -> str | None:
    if config.get_user_id:
        return config.get_user_id(request)
    return None

def _get_user_agent(request: Request, config: Config) -> str | None:
    if config.get_user_agent:
        return config.get_user_agent(request)
    if 'user-agent' in request.headers:
        return request.headers['user-agent']
    elif 'User-Agent' in request.headers:
        return request.headers['User-Agent']
    return None
