from setuptools import setup

long_description = open("README.md").read()

setup(
    name="fastapi-analytics",
    version="1.0.9",
    author="Tom Draper",
    author_email="tomjdraper1@gmail.com",
    license="GPL-3.0-or-later",
    description="Monitoring and analytics for FastAPI applications.",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/tom-draper/api-analytics",
    key_words="analytics api dashboard fastapi middleware",
    install_requires=['fastapi', 'requests'],
    packages=["api_analytics"],
    python_requires=">=3.6",
)