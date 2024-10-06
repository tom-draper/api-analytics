import logging
import subprocess
import time

import pytest
import requests
from recordings import store_recording

APP_DIR = "./tests/integration/tornado"
HOST = "127.0.0.1"
PORT = "8000"
BASE_URL = f"http://{HOST}:{PORT}"


@pytest.fixture(scope="session", autouse=True)
def start_server():
    # Start FastAPI server as a subprocess
    process = subprocess.Popen(
        ["python", "app.py"],
        cwd=APP_DIR,
    )

    # Wait for the server to start
    time.sleep(2)

    yield  # Allow tests to run

    process.terminate()
    process.wait()


def test_response():
    response = requests.get(BASE_URL)
    assert response.status_code == 200
    assert response.json() == {"message": "Hello World!"}


def test_speed():
    valid_speed(BASE_URL, "tornado")


def test_alt_log_speed():
    swap_log_level()
    logging_level_name = get_logging_level_name()
    valid_speed(BASE_URL, f"tornado (with logging level {logging_level_name})")
    swap_log_level()


def valid_speed(url: str, label: str | None = None):
    response_times = [_timed_request(url) for _ in range(100)]
    average_response_time_ms = (sum(response_times) / len(response_times)) * 1000
    print(f"Average response time: {average_response_time_ms:.4f} ms")
    store_recording(f"{average_response_time_ms:.4f}ms", label)
    assert average_response_time_ms < 15


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


def test_log():
    response = requests.get(BASE_URL)
    assert response.status_code == 200
    assert response.json() == {"message": "Hello World!"}

    time.sleep(5)

    response = requests.get(BASE_URL)
    assert response.status_code == 200
    assert response.json() == {"message": "Hello World!"}

    time.sleep(60)

    response = requests.get(BASE_URL)
    assert response.status_code == 200
    assert response.json() == {"message": "Hello World!"}
