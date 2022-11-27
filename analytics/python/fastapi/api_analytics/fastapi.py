import threading
from time import time

from starlette.middleware.base import (BaseHTTPMiddleware,
                                       RequestResponseEndpoint)
from starlette.requests import Request
from starlette.responses import Response
from starlette.types import ASGIApp

from api_analytics.core import log_request


class Analytics(BaseHTTPMiddleware):
    def __init__(self, app: ASGIApp, api_key: str) -> None:
        super().__init__(app)
        self.api_key = api_key

    async def dispatch(self, request: Request, call_next: RequestResponseEndpoint) -> Response:
        start = time()
        response: Response = await call_next(request)
        elapsed = time() - start
        
        json = {
            'api_key': self.api_key,
            'hostname': request.url.hostname,
            'path': request.url.path,
            'user_agent': request.headers['user-agent'],
            'method': request.method,
            'status': int(response.status_code),
            'response_time': int(elapsed * 1000),
            'framework': 'FastAPI'
        }
        threading.Thread(target=log_request, args=(json,)).start()
        return response
