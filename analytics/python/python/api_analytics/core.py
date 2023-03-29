import requests
import threading
from datetime import datetime


_requests = []
_last_posted = datetime.now()


def post_requests(requests_data: list[dict]):
    requests.post('http://213.168.248.206/api/log-request',
                  json=requests_data, timeout=5)


def log_request(request_data: dict):
    global _requests, _last_posted
    _requests.append(request_data)
    now = datetime.now()
    if (now - _last_posted).total_seconds() > 60.0:
        threading.Thread(target=post_requests, args=(_requests,)).start()
        _requests = []
        _last_posted = now
