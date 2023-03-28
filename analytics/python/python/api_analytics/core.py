import requests
import threading
from datetime import datetime


_requests = []
last_posted = datetime.now()


def post(data: dict):
    requests.post('http://213.168.248.206/api/log-request',
                  json=data, timeout=5)


def log_request(data: dict):
    _requests.push(data)
    now = datetime.now()
    if (now - last_posted).total_seconds() > 60.0:
        threading.Thread(target=post, args=(_requests,)).start()
        _requests = []
        last_posted = now
