
import threading
from time import time
import requests
from starlette.middleware.base import (BaseHTTPMiddleware,
                                       RequestResponseEndpoint)
from starlette.requests import Request
from starlette.responses import Response
from starlette.types import ASGIApp


class Analytics(BaseHTTPMiddleware):
    def __init__(self, app: ASGIApp, api_key: str) -> None:
        super().__init__(app)
        self.api_key = api_key

        self.method_map = {
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
    
    def log_request(self, json: dict):
        requests.post('http://localhost:8080/request', json=json)

    async def dispatch(self, request: Request, call_next: RequestResponseEndpoint) -> Response:
        start = time()
        response = await call_next(request)
        elapsed = time() - start
                
        json = {
            'api_key': self.api_key,
            'path': request.url.path,
            'user_agent': request.headers['user-agent'],
            'method': self.method_map[request.method],
            'status': int(response.status_code),
            'response_time': int(elapsed * 1000)
        }
        threading.Thread(target=self.log_request, args=(json,)).start()
        return response
