"""Proxy interface to the aiograpi client.

This module wraps the aiograpi client in a ClientProxy instance that is slightly easier to use.

See https://github.com/subzeroid/aiograpi.
"""

import asyncio
import hashlib
import logging
import os
from aiograpi import Client  # type: ignore[import-untyped]
from aiograpi.types import Account  # type: ignore[import-untyped]
from aiograpi.exceptions import LoginRequired, UserNotFound  # type: ignore[import-untyped]
from pathlib import Path
from typing import List, Optional, Tuple
from .types import AccountDict, InstagramUser

# Shared lock for `get_client()`.
LOCK = asyncio.Lock()

# Directory where to save the client's data.
PERSISTENCE_DIR = Path("/") / "mnt" / "instagram"

# Custom log handler for use with Client and ClientProxy.
log_handler = logging.StreamHandler()
log_handler.setFormatter(
    logging.Formatter(
        fmt="[%(asctime)s | %(levelname)s] %(funcName)s: %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )
)
logger = logging.getLogger("instagram_api")
logger.setLevel(logging.DEBUG)
logger.addHandler(log_handler)


class ClientProxy:
    """
    Proxy class for `aiograpi.Client`.
    """

    account: Account
    cl: Client = None

    def __init__(self, client: Client, account: Account) -> None:
        self.account = account
        self.cl = client

    async def get_account(self) -> AccountDict:
        return AccountDict(
            {
                "biography": self.account.biography or "",
                "fullName": self.account.full_name,
                "handler": self.account.username,
                "id": self.account.pk,
                "pictureURL": str(self.account.profile_pic_url),
            }
        )

    async def get_user(
        self, user_id: Optional[int] = None, handler: Optional[str] = None
    ) -> InstagramUser | None:
        try:
            if user_id:
                user = await self.cl.user_info(user_id)
            elif handler:
                user = await self.cl.user_info_by_username(handler)
            else:
                raise ValueError("Specify either the user_id or the handler")
        except UserNotFound:
            return None

        return InstagramUser(user)

    async def get_user_followers(
        self, user_id: int, next_cursor: Optional[str] = None
    ) -> Tuple[List[InstagramUser], str | None]:
        followers, next = await self.cl.user_followers_gql_chunk(
            user_id, end_cursor=next_cursor or None
        )

        next = str(next) if next else None

        return [InstagramUser(u) for u in followers], next

    async def get_user_following(
        self, user_id: int, next_cursor: Optional[str] = None
    ) -> Tuple[List[InstagramUser], str | None]:
        following, next = await self.cl.user_following_gql_chunk(
            user_id, end_cursor=next_cursor or None
        )

        next = str(next) if next else None

        return [InstagramUser(u) for u in following], next

    @classmethod
    async def get(cls, force: bool = False):
        """
        Return a ClientProxy instance. The instance is a singleton unless `force` is passed.
        """
        global CLIENT
        async with LOCK:
            if force or CLIENT is None:
                CLIENT = await init_client()

            return CLIENT

    @classmethod
    def reset(cls) -> None:
        """
        Ugly synchronous function that resets the singleton, forcing a new login when `get()` is called again.
        """
        global CLIENT
        CLIENT = None


def get_persistence_dir() -> Path:
    """
    Ensure the persistence folder exists and return it.
    """
    if not PERSISTENCE_DIR.exists():
        PERSISTENCE_DIR.mkdir(parents=True, exist_ok=True)

    return PERSISTENCE_DIR


async def init_client() -> ClientProxy:
    """
    Connect to Instagram using a session.
    https://subzeroid.github.io/aiograpi/usage-guide/best-practices.html#use-sessions
    """
    global ACCOUNT

    user, passw = os.environ.get("IG_EMAIL"), os.environ.get("IG_PASSWORD")
    if not (user and passw):
        raise Exception(
            "Instagram credentials not found, please set both IG_EMAIL and IG_PASSWORD environment variables!"
        )

    password_login, session_login = False, False
    session_hash = hashlib.md5(user.encode()).hexdigest()

    cl = Client(
        delay_range=[
            10,
            15,
        ]
    )
    cl.logger = cl.request_logger = logger

    session_file = get_persistence_dir() / f"session.{session_hash}.json"
    if session_file.exists():
        logger.info("Found existing session in %s", session_file)
        session = cl.load_settings(session_file)
    else:
        session = None

    if session:
        try:
            cl.set_settings(session)
            await cl.login(user, passw)

            try:
                ACCOUNT = await cl.account_info()
            except LoginRequired:
                logger.info("Instagram session is invalid, logging in again")

                old_session = cl.get_settings()

                cl.set_settings({})
                cl.set_uuids(old_session["uuids"])

                await cl.login(user, passw)
            session_login = True
        except Exception as e:
            logger.info("Instagram session is invalid: %s" % e)

    if not session_login:
        try:
            logger.info("Starting Instagram login")
            if await cl.login(user, passw):
                password_login = True
        except Exception as e:
            logger.info("Instagram credentials are invalid: %s" % e)

    if not password_login and not session_login:
        raise Exception("Could not initialise Instagram session")

    cl.dump_settings(session_file)

    # Workaround for https://github.com/subzeroid/aiograpi/issues/90
    cl.graphql._set_client()

    if not ACCOUNT:
        ACCOUNT = await cl.account_info()

    return ClientProxy(cl, ACCOUNT)


# Singletons.
CLIENT: ClientProxy | None = None
ACCOUNT: Account | None = None
