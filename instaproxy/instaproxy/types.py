"""Types proxied to the aiograpi library.

This module helps with exposing and mitigating breaking changes or incompatibilities
with aiograpi because its version is **not** pinned in the requirements file.
The aiograpi module should be used at its latest version at all times, because Instagram
could update their APIs or implement new anti-bot features at any moment, making this
whole program useless.

See https://github.com/subzeroid/aiograpi.
"""

from aiograpi.types import UserShort  # type: ignore[import-untyped]
from typing import TypedDict


class AccountDict(TypedDict):
    """
    JSON representation of an aiograpi.Account.
    """

    biography: str
    fullName: str
    handler: str
    id: int
    pictureURL: str


class InstagramUserDict(TypedDict):
    """
    JSON representation of an InstagramUser.
    """

    fullName: str
    handler: str
    id: int
    pictureURL: str


class InstagramUser:
    """
    Proxy class for `aiograpi` User model.
    """

    full_name: str
    handler: str
    pic_url: str
    user_id: int

    def __init__(self, api_user: UserShort):
        """
        Instantiate object using values of the provided aiograpi User instance.

        Parameters:
        -----------
        api_user: aiograpi.UserShort
        """
        self.full_name = api_user.full_name
        self.handler = api_user.username
        self.pic_url = str(api_user.profile_pic_url)
        self.user_id = int(api_user.pk)

    def to_dict(self) -> InstagramUserDict:
        return InstagramUserDict(
            {
                "fullName": self.full_name,
                "handler": self.handler,
                "id": self.user_id,
                "pictureURL": self.pic_url,
            }
        )
