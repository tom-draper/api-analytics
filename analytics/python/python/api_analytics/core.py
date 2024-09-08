import logging
import threading
from datetime import datetime
from types import NoneType
from typing import Dict, List, Union

import requests

DEFAULT_SERVER_URL = "https://www.apianalytics-server.com"

_requests = []
_last_posted = datetime.now()

logger = logging.getLogger("api_analytics")
logger.setLevel(logging.DEBUG)


def log_request(
    api_key: str,
    request_data: Dict,
    framework: str,
    privacy_level: int,
    server_url: str,
):
    logger.debug(f"Logging request: {request_data}")
    if api_key == "" or api_key is None:
        logging.debug("Aborting log request: API key is not set.")
        return

    global _requests, _last_posted
    _requests.append(request_data)
    now = datetime.now()
    if (now - _last_posted).total_seconds() > 60.0:
        threading.Thread(
            target=_post_requests,
            args=(api_key, _requests, framework, privacy_level, server_url),
        ).start()
        _requests = []
        _last_posted = now


def _post_requests(
    api_key: str,
    requests_data: List[Dict],
    framework: str,
    privacy_level: int,
    server_url: str,
):
    url = _endpoint_url(server_url)
    logger.debug(f"Posting {len(requests_data)} logged requests to server: {url}")
    if url is None:
        logger.debug("Aborting post to server: Server URL is not set.")
        return

    response = requests.post(
        url,
        json={
            "api_key": api_key,
            "requests": requests_data,
            "framework": framework,
            "privacy_level": privacy_level,
        },
        timeout=10,
    )
    logger.debug(f"Response from server ({response.status_code}): {response.text}")


def _endpoint_url(server_url: Union[str, NoneType]):
    if server_url is None or server_url == "":
        return server_url
    elif server_url[-1] == "/":
        return server_url + "api/log-request"
    return server_url + "/api/log-request"
