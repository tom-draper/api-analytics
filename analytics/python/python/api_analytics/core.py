import requests
import threading
from datetime import datetime


_requests = []
_last_posted = datetime.now()


def _post_requests(api_key: str, requests_data: list[dict], framework: str):
    requests.post('https://213.168.248.206/api/log-request',
                  json={
                      'api_key': api_key,
                      'requests': requests_data,
                      'framework': framework
                  }, timeout=5)


def log_request(api_key: str, request_data: dict, framework: str):
    if api_key == "" or api_key is None:
        return
    global _requests, _last_posted
    _requests.append(request_data)
    now = datetime.now()
    if (now - _last_posted).total_seconds() > 60.0:
        threading.Thread(target=_post_requests, args=(
            api_key, _requests, framework)).start()
        _requests = []
        _last_posted = now
