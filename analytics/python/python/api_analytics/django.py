from dataclasses import dataclass
from datetime import datetime
from time import time
from typing import Callable, Union

from api_analytics.core import log_request, DEFAULT_SERVER_URL
from django import conf
from django.core.handlers.wsgi import WSGIRequest
from django.http.response import HttpResponse


@dataclass
class Config:
    """
    Configuration for the Django API Analytics middleware.

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
    get_path: Callable[[WSGIRequest], Union[str, None]] = Analytics.Mappers.get_path
    get_ip_address: Callable[[WSGIRequest], Union[str, None]] = Analytics.Mappers.get_ip_address
    get_hostname: Callable[[WSGIRequest], Union[str, None]] = Analytics.Mappers.get_hostname
    get_user_agent: Callable[[WSGIRequest], Union[str, None]] = Analytics.Mappers.get_user_agent
    get_user_id: Callable[[WSGIRequest], Union[str, None]] = Analytics.Mappers.get_user_id


class Analytics:
    """API Analytics middleware for Django."""

    def __init__(self, get_response: Callable[[WSGIRequest], HttpResponse]):
        self.get_response = get_response
        self.api_key = getattr(conf.settings, "ANALYTICS_API_KEY", None)
        self.config = getattr(conf.settings, "ANALYTICS_CONFIG", Config())

    def __call__(self, request: WSGIRequest) -> HttpResponse:
        start = time()
        response = self.get_response(request)

        request_data = {
            "hostname": self.config.get_hostname(request),
            "ip_address": self._get_ip_address(request),
            "path": self.config.get_path(request),
            "user_agent": self.config.get_user_agent(request),
            "method": request.method,
            "status": response.status_code,
            "response_time": int((time() - start) * 1000),
            "user_id": self.config.get_user_id(request),
            "created_at": datetime.now().isoformat(),
        }

        log_request(
            self.api_key, 
            request_data, 
            "Django", 
            self.config.privacy_level, 
            self.config.server_url
        )
        return response
    
    class Mappers:
        @staticmethod
        def get_path(request: WSGIRequest) -> Union[str, None]:
            return request.path

        @staticmethod
        def get_ip_address(request: WSGIRequest) -> Union[str, None]:
            return request.META.get("REMOTE_ADDR"))

        @staticmethod
        def get_hostname(request: WSGIRequest) -> Union[str, None]:
            return request.get_host()
        
        @staticmethod
        def get_user_id(request: WSGIRequest) -> Union[str, None]:
            return None
        
        @staticmethod
        def get_user_agent(request: WSGIRequest) -> Union[str, None]:
            if "user-agent" in request.headers:
                return request.headers["user-agent"]
            elif "User-Agent" in request.headers:
                return request.headers["User-Agent"]
            return None

    def _get_ip_address(self, request: WSGIRequest) -> Union[str, None]:
        # If privacy_level is max, client IP address is never sent to the server
        if self.config.privacy_level >= 2:
            return None
        return self.config.get_ip_address(request)
