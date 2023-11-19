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
    
    @staticmethod
    def _get_user_agent(request: Request) -> str:
        user_agent = ''
        if 'user-agent' in request.headers:
            user_agent = request.headers['user-agent']
        elif 'User-Agent' in request.headers:
            user_agent = request.headers['User-Agent']
        return user_agent

    async def dispatch(self, request: Request, call_next: RequestResponseEndpoint) -> Response:
        start = time()
        response = await call_next(request)

        request_data = {
            'hostname': request.url.hostname,
            'ip_address': request.client.host,
            'path': request.url.path,
            'user_agent': self._get_user_agent(request),
            'method': request.method,
            'status': response.status_code,
            'response_time': int((time() - start) * 1000),
            'created_at': datetime.now().isoformat(),
        }

        log_request(self.api_key, request_data, 'FastAPI')
        return response
