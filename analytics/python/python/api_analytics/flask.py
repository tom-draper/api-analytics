from dataclasses import dataclass
from datetime import datetime
from time import time
from typing import Callable, Union

from api_analytics.core import log_request
from flask import Flask, Request, Response, request


@dataclass
class Config:
    """
    Configuration for the Flask API Analytics middleware.

    :param get_path: Optional custom mapping function that takes a request and
        returns the path stored within the request. If set, it overrides the
        default behaviour of API Analytics.
    :param get_ip_address: Optional custom mapping function that takes a request
        and returns the IP address stored within the request. If set, it
        overrides the default behaviour of API Analytics.
    :param get_hostname: Optional custom mapping function that takes a request
        and returns the hostname stored within the request. If set, it overrides
        the default behaviour of API Analytics.
    :param get_user_agent: Optional custom mapping function that takes a request
        and returns the user agent stored within the request. If set, it
        overrides the default behaviour of API Analytics.
    :param get_user_id: Optional custom mapping function that takes a request
        and returns a custom user ID stored within the request. If set, this can
        be used to track a custom user ID specific to your API such as an API
        key or client ID. If left as `None`, no custom user ID will be used, and
        user identification may rely on client IP address only (depending on
        `privacy_level`).
    :param privacy_level: Controls client identification by IP address.
        - 0: Sends client IP to the server to be stored and client location is
        inferred.
        - 1: Sends the client IP to the server only for the location to be
        inferred and stored, with the IP discarded afterwards.
        - 2: Avoids sending the client IP address to the server. Providing a
        custom `get_user_id` mapping function becomes the only method for client
        identification.
        Defaults to 0.
    """

    get_path: Union[Callable[[Request], str], None] = None
    get_ip_address: Union[Callable[[Request], str], None] = None
    get_hostname: Union[Callable[[Request], str], None] = None
    get_user_agent: Union[Callable[[Request], str], None] = None
    get_user_id: Union[Callable[[Request], str], None] = None
    privacy_level: int = 0


def add_middleware(app: Flask, api_key: str, config: Config = Config()):
    """
    Adds API Analytics middleware to the Flask app to log requests to the server.

    :param app: Flask app to attach the analytics middleware to
    :param api_key: API key for API Analytics
    :param config: Optional configuration for the middleware
    """
    start = 0.0

    @app.before_request
    def prepare():
        nonlocal start
        start = time()

    @app.after_request
    def on_finish(response: Response) -> Response:
        nonlocal start
        request_data = {
            "hostname": _get_hostname(request, config),
            "ip_address": _get_ip_address(request, config),
            "path": _get_path(request, config),
            "user_agent": _get_user_agent(request, config),
            "method": request.method,
            "status": response.status_code,
            "response_time": int((time() - start) * 1000),
            "user_id": _get_user_id(request, config),
            "created_at": datetime.now().isoformat(),
        }

        log_request(api_key, request_data, "Flask", config.privacy_level)
        return response


def _get_path(request: Request, config: Config) -> Union[str, None]:
    if config.get_path:
        return config.get_path(request)
    return request.path


def _get_ip_address(request: Request, config: Config) -> Union[str, None]:
    # If privacy_level is max, client IP address is never sent to the server
    if config.privacy_level >= 2:
        return None

    if config.get_ip_address:
        return config.get_ip_address(request)
    return request.remote_addr


def _get_hostname(request: Request, config: Config) -> Union[str, None]:
    if config.get_hostname:
        return config.get_hostname(request)
    return request.host


def _get_user_id(request: Request, config: Config) -> Union[str, None]:
    if config.get_user_id:
        return config.get_user_id(request)
    return None


def _get_user_agent(request: Request, config: Config) -> Union[str, None]:
    if config.get_user_agent:
        return config.get_user_agent(request)
    if "user-agent" in request.headers:
        return request.headers["user-agent"]
    elif "User-Agent" in request.headers:
        return request.headers["User-Agent"]
    return None
