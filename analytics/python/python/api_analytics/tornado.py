from dataclasses import dataclass
from datetime import datetime
from time import time
from typing import Callable, Union

from .core import log_request, logger, DEFAULT_SERVER_URL
from tornado.httputil import HTTPServerRequest
from tornado.web import Application, RequestHandler


class Analytics(RequestHandler):
    """API Analytics middleware for Tornado."""

    def __init__(
        self,
        app: Application,
        res: HTTPServerRequest,
        api_key: str,
        config: "Config" = None,
    ):
        super().__init__(app, res)
        self.api_key = api_key
        self.config = config or Config()
        self.start = time()

        if not self.api_key:
            logger.debug("API key is not set.")
        if not self.config.server_url:
            logger.debug("Server URL is not set.")

    def prepare(self):
        self.start = time()

    def on_finish(self):
        request_data = {
            "hostname": self.config.get_hostname(),
            "ip_address": self._get_ip_address(),
            "path": self.config.get_path(),
            "user_agent": self.config.get_user_agent(),
            "method": self.request.method,
            "status": self.get_status(),
            "response_time": int((time() - self.start) * 1000),
            "user_id": self.config.get_user_id(),
            "created_at": datetime.now().isoformat(),
        }

        log_request(
            self.api_key,
            request_data,
            "Tornado",
            self.config.privacy_level,
            self.config.server_url,
        )
        self.start = None

    class Mappers:
        @staticmethod
        def get_path(request: HTTPServerRequest) -> Union[str, None]:
            return request.path

        @staticmethod
        def get_ip_address(request: HTTPServerRequest) -> Union[str, None]:
            return request.remote_ip

        @staticmethod
        def get_hostname(request: HTTPServerRequest) -> Union[str, None]:
            return request.host

        @staticmethod
        def get_user_id(request: HTTPServerRequest) -> Union[str, None]:
            return None

        @staticmethod
        def get_user_agent(request: HTTPServerRequest) -> Union[str, None]:
            if "user-agent" in request.headers:
                return request.headers["user-agent"]
            elif "User-Agent" in request.headers:
                return request.headers["User-Agent"]
            return None

    def _get_ip_address(self) -> Union[str, None]:
        # If privacy_level is max, client IP address is never sent to the server
        if self.config.privacy_level >= 2:
            return None

        if self.config.get_ip_address:
            return self.config.get_ip_address(self.request)
        return self.request.remote_ip


@dataclass
class Config:
    """
    Configuration for the Tornado API Analytics middleware.

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
    get_path: Callable[
        [HTTPServerRequest], Union[str, None]
    ] = Analytics.Mappers.get_path
    get_ip_address: Callable[
        [HTTPServerRequest], Union[str, None]
    ] = Analytics.Mappers.get_ip_address
    get_hostname: Callable[
        [HTTPServerRequest], Union[str, None]
    ] = Analytics.Mappers.get_hostname
    get_user_agent: Callable[
        [HTTPServerRequest], Union[str, None]
    ] = Analytics.Mappers.get_user_agent
    get_user_id: Callable[
        [HTTPServerRequest], Union[str, None]
    ] = Analytics.Mappers.get_user_id
