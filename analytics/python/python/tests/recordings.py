_session_recordings = []


def get_recordings():
    return _session_recordings


def store_recording(value: any, label: str):
    _session_recordings.append({"value": value, "label": label})
