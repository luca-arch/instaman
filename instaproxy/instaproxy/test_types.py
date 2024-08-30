from pydantic import HttpUrl
import pytest
from types import SimpleNamespace
from .types import InstagramUser


def get_user():
    api_user = SimpleNamespace(
        full_name="Test User",
        pk="123",
        profile_pic_url=HttpUrl("http://www.example.com/1.jpg"),
        username="test_user",
    )

    return InstagramUser(api_user)


@pytest.mark.anyio
async def test_attrs():
    user = get_user()

    assert user.full_name == "Test User"
    assert user.handler == "test_user"
    assert user.pic_url == "http://www.example.com/1.jpg"
    assert user.user_id == 123


@pytest.mark.anyio
async def test_to_dict():
    user = get_user()

    assert user.to_dict() == {
        "fullName": user.full_name,
        "handler": user.handler,
        "id": user.user_id,
        "pictureURL": user.pic_url,
    }
