import subprocess
import time

import pytest
from utils import valid_alt_log_speed, valid_response, valid_log, valid_speed

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
    valid_response(BASE_URL)


def test_speed():
    valid_speed(BASE_URL, 'tornado')


def test_alt_log_speed():
    valid_alt_log_speed(BASE_URL, 'tornado')


@pytest.mark.full
def test_log():
    valid_log(BASE_URL)
