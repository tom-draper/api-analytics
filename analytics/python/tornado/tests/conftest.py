import pytest
from analytics.python.fastapi.tests.recordings import get_recordings


# This hook runs after all tests have completed
def pytest_sessionfinish(session, exitstatus):
    recordings = get_recordings()
    for recording in recordings:
        print(f"{recording['label']}: {recording['value']}")


def pytest_addoption(parser):
    # Add the --full option to the command line
    parser.addoption(
        "--full",
        action="store_true",
        default=False,
        help="Run all tests (default skips those marked with @pytest.mark.full)",
    )


def pytest_collection_modifyitems(config, items):
    # If --full is not passed, skip tests marked as "full"
    if not config.getoption("--full"):
        skip_full = pytest.mark.skip(
            reason="Running quick tests, use --full to run all"
        )
        for item in items:
            if "full" in item.keywords:
                item.add_marker(skip_full)
