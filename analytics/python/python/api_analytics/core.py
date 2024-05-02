import threading
from datetime import datetime
from typing import Dict, List

import requests


DEFAULT_SERVER_URL = 'https://www.apianalytics-server.com'

_requests = []
_last_posted = datetime.now()


def _post_requests(
    api_key: str, requests_data: List[Dict], framework: str, privacy_level: int, server_url: str
):
    requests.post(
        endpoint_url(server_url),
        json={
            "api_key": api_key,
            "requests": requests_data,
            "framework": framework,
            "privacy_level": privacy_level,
        },
        timeout=10,
    )


def log_request(api_key: str, request_data: Dict, framework: str, privacy_level: int, server_url: str):
    if api_key == "" or api_key is None:
        return
    global _requests, _last_posted
    _requests.append(request_data)
    now = datetime.now()
    if (now - _last_posted).total_seconds() > 60.0:
        threading.Thread(
            target=_post_requests, args=(api_key, _requests, framework, privacy_level, server_url)
        ).start()
        _requests = []
        _last_posted = now


def endpoint_url(server_url: str):
    if server_url is None or server_url == "":
        return server_url
    elif server_url[-1] == "/":
        return server_url + "api/log-request"
    return server_url + "/api/log_request"