from setuptools import setup

long_description = open("README.md").read()

setup(
    name="api-analytics",
    version="1.0.5",
    author="Tom Draper",
    author_email="tomjdraper1@gmail.com",
    license="MIT",
    description="Monitoring and analytics for Python API frameworks.",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/tom-draper/api-analytics",
    key_words="analytics api dashboard fastapi flask",
    install_requires=['fastapi', 'Flask', 'requests'],
    packages=["api_analytics"],
    python_requires=">=3.6",
)