import requests

def log_request(data: dict):
    requests.post('https://api-analytics-server.vercel.app/api/log-request', json=data, timeout=5)
