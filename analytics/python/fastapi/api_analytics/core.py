import requests


def log_request(data: dict):
    # requests.post(
    # 'https://api-analytics-server.vercel.app/api/log-request', json=data, timeout=5)
    requests.post('http://213.168.248.206/api/log-request',
                  json=data, timeout=5)
