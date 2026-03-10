import logging
import threading
from datetime import datetime
from typing import Dict, List

import requests

DEFAULT_SERVER_URL = "https://www.apianalytics-server.com"

_lock = threading.Lock()
_requests: Dict[str, List[Dict]] = {}
_last_posted: Dict[str, datetime] = {}

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
    if not api_key:
        logger.debug("Aborting log request: API key is not set.")
        return

    requests_to_post = None
    with _lock:
        if api_key not in _requests:
            _requests[api_key] = []
            _last_posted[api_key] = datetime.now()

        _requests[api_key].append(request_data)
        now = datetime.now()
        if (now - _last_posted[api_key]).total_seconds() > 60.0:
            requests_to_post = list(_requests[api_key])
            _requests[api_key] = []
            _last_posted[api_key] = now

    if requests_to_post:
        threading.Thread(
            target=_post_requests,
            args=(api_key, requests_to_post, framework, privacy_level, server_url),
        ).start()


def _post_requests(
    api_key: str,
    requests_data: List[Dict],
    framework: str,
    privacy_level: int,
    server_url: str,
):
    url = _endpoint_url(server_url)
    logger.debug(f"Posting {len(requests_data)} logged requests to server: {url}")

    try:
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
    except Exception as e:
        logger.debug(f"Failed to post logs: {e}")


def _endpoint_url(server_url: str) -> str:
    if not server_url:
        return DEFAULT_SERVER_URL + "/api/log-request"
    if server_url.endswith("/"):
        return server_url + "api/log-request"
    return server_url + "/api/log-request"
