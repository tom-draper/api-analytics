from setuptools import setup

long_description = open("README.md").read()

setup(
    name="api-analytics",
    version="1.1.3",
    author="Tom Draper",
    author_email="tomjdraper1@gmail.com",
    license="MIT",
    description="Monitoring and analytics for Python API frameworks.",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/tom-draper/api-analytics",
    key_words="analytics api dashboard django fastapi flask tornado middleware",
    install_requires=['Django', 'fastapi', 'Flask', 'tornado', 'requests'],
    packages=["api_analytics"],
    python_requires=">=3.6",
)