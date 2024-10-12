import logging
import time

import requests
from recordings import store_recording


def valid_log(url: str):
    response = requests.get(url)
    assert response.status_code == 200
    assert response.json() == {"message": "Hello World!"}

    time.sleep(5)

    response = requests.get(url)
    assert response.status_code == 200
    assert response.json() == {"message": "Hello World!"}

    time.sleep(60)

    response = requests.get(url)
    assert response.status_code == 200
    assert response.json() == {"message": "Hello World!"}


def valid_response(url: str):
    response = requests.get(url)
    assert response.status_code == 200
    assert response.json() == {"message": "Hello World!"}


def valid_speed(url: str, label: str | None = None):
    response_times = [_timed_request(url) for _ in range(100)]
    average_response_time_ms = (sum(response_times) / len(response_times)) * 1000
    print(f"Average response time: {average_response_time_ms:.4f} ms")
    store_recording(f"{average_response_time_ms:.4f}ms", label)
    assert average_response_time_ms < 15


def valid_alt_log_speed(url: str, label: str | None = None):
    swap_log_level()
    logging_level_name = get_logging_level_name()
    valid_speed(url, f"{label} (with logging level {logging_level_name})")
    swap_log_level()


def get_logging_level_name():
    logger = logging.getLogger("api_analytics")
    return logging.getLevelName(logger.getEffectiveLevel())


def swap_log_level():
    logger = logging.getLogger("api_analytics")
    logging_level = logger.getEffectiveLevel()
    if logging_level <= logging.DEBUG:
        logger.setLevel(logging.INFO)
    else:
        logger.setLevel(logging.DEBUG)


def _timed_request(url: str):
    start = time.time()
    _ = requests.get(url)
    end = time.time()
    return end - start
