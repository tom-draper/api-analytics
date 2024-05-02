from dataclasses import dataclass
from datetime import datetime
from time import time
from typing import Callable, Union

from api_analytics.core import log_request, DEFAULT_SERVER_URL
from flask import Flask, Request, Response, request


@dataclass
class Config:
    """
    Configuration for the Flask API Analytics middleware.

    :param privacy_level: Controls client identification by IP address.
        - 0: Sends client IP to the server to be stored and client location is
        inferred.
        - 1: Sends the client IP to the server only for the location to be
        inferred and stored, with the IP discarded afterwards.
        - 2: Avoids sending the client IP address to the server. Providing a
        custom `get_user_id` mapping function becomes the only method for client
        identification.
        Defaults to 0.
    :param server_url: For self-hosting. Points to the public server url to post 
        requests to.
    :param get_path: Mapping function that takes a request and returns the path 
        stored within the request. Assigning a value will override the default 
        behavior.
    :param get_ip_address: Mapping function that takes a request and returns the 
        IP address stored within the request. Assigning a value will override the 
        default behavior.
    :param get_hostname: Mapping function that takes a request and returns the 
        hostname stored within the request. Assigning a value will override the 
        default behavior.
    :param get_user_agent: Mapping function that takes a request and returns the 
        user agent stored within the request. Assigning a value will override the 
        default behavior.
    :param get_user_id: Mapping function that takes a request and returns a 
        custom user ID stored within the request. Always returns None by default. 
        Assigning a value allows for tracking a custom user ID specific to your API 
        such as an API key or client ID. If left as the default value, user 
        identification may rely on client IP address only (depending on
        `privacy_level`).
    """

    privacy_level: int = 0
    server_url: str = DEFAULT_SERVER_URL
    get_path: Callable[[Request], Union[str, None]] = Mappers.get_path
    get_ip_address: Callable[[Request], Union[str, None]] = Mappers.get_ip_address
    get_hostname: Callable[[Request], Union[str, None]] = Mappers.get_hostname
    get_user_agent: Callable[[Request], Union[str, None]] = Mappers.get_user_agent
    get_user_id: Callable[[Request], Union[str, None]] = Mappers.get_user_id


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
            "hostname": config.get_hostname(request, config),
            "ip_address": _get_ip_address(request, config),
            "path": config.get_path(request, config),
            "user_agent": config.get_user_agent(request, config),
            "method": request.method,
            "status": response.status_code,
            "response_time": int((time() - start) * 1000),
            "user_id": config.get_user_id(request, config),
            "created_at": datetime.now().isoformat(),
        }

        log_request(
            api_key, 
            request_data, 
            "Flask", 
            config.privacy_level, 
            config.server_url
        )
        return response

class Mappers:
    @staticmethod
    def get_path(request: Request) -> Union[str, None]:
        return request.path

    @staticmethod
    def get_ip_address(request: Request) -> Union[str, None]:
        return request.remote_addr

    @staticmethod
    def get_hostname(request: Request) -> Union[str, None]:
        return request.url.hostname
    
    @staticmethod
    def get_user_id(request: Request) -> Union[str, None]:
        return None
    
    @staticmethod
    def get_user_agent(request: Request) -> Union[str, None]:
        if "user-agent" in request.headers:
            return request.headers["user-agent"]
        elif "User-Agent" in request.headers:
            return request.headers["User-Agent"]
        return None

def _get_ip_address(request: Request, config: Config) -> Union[str, None]:
    # If privacy_level is max, client IP address is never sent to the server
    if config.privacy_level >= 2:
        return None
    return config.get_ip_address(request)
