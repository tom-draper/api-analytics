import time

import requests


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


def valid_speed(url: str):
    response_times = [timed_request() for _ in range(100)]
    average_response_time = sum(response_times) / len(response_times)
    print(f"Average response time: {average_response_time*1000:.4f} ms")
    assert average_response_time < 0.01


def timed_request(url: str):
    start = time.time()
    _ = requests.get(url)
    end = time.time()
    return end - start
