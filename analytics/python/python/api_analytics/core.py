import atexit
import logging
import threading
import time
from datetime import datetime
from typing import Dict, List, Tuple

import requests

DEFAULT_SERVER_URL = "https://www.apianalytics-server.com"
FLUSH_INTERVAL_SECONDS = 60.0

_lock = threading.Lock()
_requests: Dict[str, List[Dict]] = {}
_last_posted: Dict[str, datetime] = {}
# Per-key posting metadata (framework, privacy_level, server_url), recorded on
# each log call so the background flusher and the atexit hook can post a batch
# without it being threaded through to them.
_meta: Dict[str, Tuple[str, int, str]] = {}
_flusher_started = False
_flusher_lock = threading.Lock()

logger = logging.getLogger("api_analytics")
logger.setLevel(logging.DEBUG)


def log_request(
    api_key: str,
    request_data: Dict,
    framework: str,
    privacy_level: int,
    server_url: str,
):
    logger.debug(f"Logging request: {request_data}")
    if not api_key:
        logger.debug("Aborting log request: API key is not set.")
        return

    _ensure_flusher()

    requests_to_post = None
    with _lock:
        if api_key not in _requests:
            _requests[api_key] = []
            _last_posted[api_key] = datetime.now()

        _requests[api_key].append(request_data)
        _meta[api_key] = (framework, privacy_level, server_url)
        now = datetime.now()
        if (now - _last_posted[api_key]).total_seconds() > FLUSH_INTERVAL_SECONDS:
            requests_to_post = list(_requests[api_key])
            _requests[api_key] = []
            _last_posted[api_key] = now

    if requests_to_post:
        threading.Thread(
            target=_post_or_requeue,
            args=(api_key, requests_to_post, framework, privacy_level, server_url),
        ).start()


def flush() -> None:
    """Post every buffered request for every API key.

    Called periodically by the background flusher and on interpreter exit, so a
    partial batch is not held indefinitely when traffic goes idle or lost when
    the process shuts down. Safe to call directly from an application's own
    shutdown hook.
    """
    batches = []
    with _lock:
        for api_key, buffered in _requests.items():
            if buffered and api_key in _meta:
                batches.append((api_key, buffered, _meta[api_key]))
                _requests[api_key] = []
                _last_posted[api_key] = datetime.now()

    for api_key, buffered, (framework, privacy_level, server_url) in batches:
        _post_or_requeue(api_key, buffered, framework, privacy_level, server_url)


def _ensure_flusher():
    # Lazily start a single daemon thread that flushes buffered requests every
    # FLUSH_INTERVAL_SECONDS, and register an atexit hook to flush on shutdown.
    # Started on the first logged request rather than at import so it runs in the
    # worker process of a forking server.
    global _flusher_started
    if _flusher_started:
        return
    with _flusher_lock:
        if _flusher_started:
            return
        _flusher_started = True

    def _run():
        while True:
            time.sleep(FLUSH_INTERVAL_SECONDS)
            try:
                flush()
            except Exception as e:
                logger.debug(f"Background flush failed: {e}")

    threading.Thread(target=_run, daemon=True).start()
    atexit.register(flush)


def _post_requests(
    api_key: str,
    requests_data: List[Dict],
    framework: str,
    privacy_level: int,
    server_url: str,
) -> bool:
    url = _endpoint_url(server_url)
    logger.debug(f"Posting {len(requests_data)} logged requests to server: {url}")

    try:
        response = requests.post(
            url,
            json={
                "api_key": api_key,
                "requests": requests_data,
                "framework": framework,
                "privacy_level": privacy_level,
            },
            timeout=10,
        )
        logger.debug(f"Response from server ({response.status_code}): {response.text}")
        return 200 <= response.status_code < 300
    except Exception as e:
        logger.debug(f"Failed to post logs: {e}")
        return False


def _post_or_requeue(
    api_key: str,
    requests_data: List[Dict],
    framework: str,
    privacy_level: int,
    server_url: str,
):
    """Upload a batch, restoring it ahead of newer requests on failure."""
    if _post_requests(api_key, requests_data, framework, privacy_level, server_url):
        return

    with _lock:
        _requests[api_key] = requests_data + _requests.get(api_key, [])


def _endpoint_url(server_url: str) -> str:
    if not server_url:
        return DEFAULT_SERVER_URL + "/api/log-request"
    if server_url.endswith("/"):
        return server_url + "api/log-request"
    return server_url + "/api/log-request"
