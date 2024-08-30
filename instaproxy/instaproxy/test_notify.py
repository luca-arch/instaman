import os
import pytest

os.environ["TG_BOT_TOKEN"] = "1234"
os.environ["TG_CHANNEL"] = "-9876"

from . import notify  # noqa: E402


@pytest.mark.anyio
async def test_enqueue_exception():
    e0 = ValueError("value error")
    notify.enqueue_exception(e0)

    e1 = AttributeError("attribute error")
    notify.enqueue_exception(e1)

    assert notify.QUEUES.keys() == {
        "ValueError",
        "AttributeError",
    }

    assert notify.QUEUES["ValueError"].qsize() == 1
    assert notify.QUEUES["ValueError"].get_nowait() == e0

    assert notify.QUEUES["AttributeError"].qsize() == 1
    assert notify.QUEUES["AttributeError"].get_nowait() == e1

    e2 = ValueError("value error 2")
    e3 = ValueError("value error 3")

    notify.enqueue_exception(e2)
    notify.enqueue_exception(e3)

    assert notify.QUEUES["ValueError"].qsize() == 2
    assert notify.QUEUES["ValueError"].get_nowait() == e2
    assert notify.QUEUES["ValueError"].get_nowait() == e3


@pytest.mark.anyio
async def test_get_message_text():
    errors = [
        TypeError("type error 1"),
        TypeError("type error 2"),
    ]

    msg_lines = notify.get_message_text(errors).split("\n")

    assert msg_lines == [
        """💥 **Instaproxy error (TypeError): type error 1**""",
        "",
        "```",
        "TypeError: type error 1",
        "",
        "```",
        "",
        "This error occurred 2 times, only the first stack trace is attached",
    ]
