from datetime import datetime

import api_analytics.core as core


def setup_function():
    core._requests.clear()
    core._last_posted.clear()
    core._meta.clear()


def test_flush_requeues_batch_after_server_error(monkeypatch):
    api_key = "test-key"
    request_data = {"path": "/"}
    core._requests[api_key] = [request_data]
    core._meta[api_key] = ("FastAPI", 0, "https://example.test")
    core._last_posted[api_key] = datetime.now()

    class Response:
        status_code = 503
        text = "temporarily unavailable"

    monkeypatch.setattr(core.requests, "post", lambda *args, **kwargs: Response())

    core.flush()

    assert core._requests[api_key] == [request_data]
