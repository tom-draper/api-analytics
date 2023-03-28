from datetime import datetime
from time import time

from api_analytics.core import log_request
from starlette.middleware.base import (BaseHTTPMiddleware,
                                       RequestResponseEndpoint)
from starlette.requests import Request
from starlette.responses import Response
from starlette.types import ASGIApp


class Analytics(BaseHTTPMiddleware):
    def __init__(self, app: ASGIApp, api_key: str):
        super().__init__(app)
        self.api_key = api_key

    async def dispatch(self, request: Request, call_next: RequestResponseEndpoint) -> Response:
        start = time()
        response = await call_next(request)

        data = {
            'api_key': self.api_key,
            'hostname': request.url.hostname,
            'ip_address': request.client.host,
            'path': request.url.path,
            'user_agent': request.headers['user-agent'],
            'method': request.method,
            'status': response.status_code,
            'framework': 'FastAPI',
            'response_time': int((time() - start) * 1000),
            'created_at': datetime.now()
        }

        log_request(data)
        return response
