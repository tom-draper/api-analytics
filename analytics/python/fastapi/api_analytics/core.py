import requests

method_map = {
    'GET': 0,
    'POST': 1,
    'PUT': 2,
    'PATCH': 3,
    'DELETE': 4,
    'OPTIONS': 5,
    'CONNECT': 6,
    'HEAD': 7,
    'TRACE': 8,
}

def log_request(json: dict):
    json['method'] = method_map[json['method']]
    requests.post('https://api-analytics-server.vercel.app/api/log-request', json=json, timeout=5)
