from dataclasses import dataclass
from datetime import datetime
from time import time
from typing import Callable, Union

from api_analytics.core import log_request
from django import conf
from django.core.handlers.wsgi import WSGIRequest
from django.http.response import HttpResponse


@dataclass
class Config:
    """
    Configuration for the Django API Analytics middleware.

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

    get_path: Union[Callable[[WSGIRequest], str], None] = None
    get_ip_address: Union[Callable[[WSGIRequest], str], None] = None
    get_hostname: Union[Callable[[WSGIRequest], str], None] = None
    get_user_agent: Union[Callable[[WSGIRequest], str], None] = None
    get_user_id: Union[Callable[[WSGIRequest], str], None] = None
    privacy_level: int = 0


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
            "hostname": self._get_hostname(request),
            "ip_address": self._get_ip_address(request),
            "path": self._get_path(request),
            "user_agent": self._get_user_agent(request),
            "method": request.method,
            "status": response.status_code,
            "response_time": int((time() - start) * 1000),
            "user_id": self._get_user_id(request),
            "created_at": datetime.now().isoformat(),
        }

        log_request(self.api_key, request_data, "Django")
        return response

    def _get_path(self, request: WSGIRequest) -> Union[str, None]:
        if self.config.get_path:
            return self.config.get_path(request)
        return request.path

    def _get_ip_address(self, request: WSGIRequest) -> Union[str, None]:
        # If privacy_level is max, client IP address is never sent to the server
        if self.config.privacy_level >= 2:
            return None

        if self.config.get_ip_address:
            return self.config.get_ip_address(request)
        return request.META.get("REMOTE_ADDR")

    def _get_hostname(self, request: WSGIRequest) -> Union[str, None]:
        if self.config.get_hostname:
            return self.config.get_hostname(request)
        return request.get_host()

    def _get_user_id(self, request: WSGIRequest) -> Union[str, None]:
        if self.config.get_user_id:
            return self.config.get_user_id(request)
        return None

    def _get_user_agent(self, request: WSGIRequest) -> Union[str, None]:
        if self.config.get_user_agent:
            return self.config.get_user_agent(request)
        elif "user-agent" in request.headers:
            return request.headers["user-agent"]
        elif "User-Agent" in request.headers:
            return request.headers["User-Agent"]
        return None
