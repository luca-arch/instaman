"""Rest API server.

This modules spins up a Fastapi instance that provides several endpoints to connect
with Instagram.
"""

from aiograpi.exceptions import (  # type: ignore[import-untyped]
    ChallengeRequired,
    ClientLoginRequired,
    ClientUnauthorizedError,
    LoginRequired,
    ProxyError,
)
from contextlib import asynccontextmanager
from fastapi import FastAPI, Request
from fastapi.exceptions import HTTPException
from fastapi.responses import JSONResponse
from typing import Optional
from .cache import CacheWithTTL
from .client import ClientProxy
from .notify import enqueue_exception, start_notifier
from .types import AccountDict, InstagramUserDict


# List of aiograpi exceptions for which the client should be re-instantiated.
AUTH_EXCEPTIONS = (
    ClientLoginRequired,
    ClientUnauthorizedError,
    ProxyError,
    LoginRequired,
)


# List of aiograpi exceptions that should be notified (the Instagram account likely requires human interaction).
NOTIFY_EXCEPTIONS = (ChallengeRequired,)


@asynccontextmanager
async def lifespan(*args, **kwargs):
    """
    Lifespan event for sending Telegram notifications.

    See https://fastapi.tiangolo.com/advanced/events/.
    """
    task = await start_notifier()

    yield

    task.cancel()


app = FastAPI(lifespan=lifespan)


@app.get("/me")
async def get_account() -> AccountDict:
    """Return information of the account currently logged in with the client.

    Returns
    -------
    AccountDict
        the account's information.
    """
    client = await ClientProxy.get()

    return await client.get_account()


@app.get("/account/{handler}")
@CacheWithTTL.decorate(ttl=60 * 10)
async def get_user(handler: str) -> InstagramUserDict:
    """Return information of the user given their @handler.

    Parameters
    ----------
    handler: string
        The user's handler (with/without the @ prefix).

    Returns
    -------
    InstagramUserDict
        the user's information.
    """
    handler = handler.strip("@")

    client = await ClientProxy.get()

    user = await client.get_user(handler=handler)
    if not user:
        raise HTTPException(404, detail=f"User {handler} does not exist")

    return user.to_dict()


@app.get("/account-id/{user_id}")
@CacheWithTTL.decorate(ttl=60 * 10)
async def get_user_by_id(user_id: int) -> InstagramUserDict:
    """Return information of the user given their ID.

    Parameters
    ----------
    user_id: string
        The user's ID.

    Returns
    -------
    InstagramUserDict
        the user's information.
    """
    client = await ClientProxy.get()

    user = await client.get_user(user_id=user_id)
    if not user:
        raise HTTPException(404, detail=f"User with ID {user_id} does not exist")

    return user.to_dict()


@app.get("/followers/{user_id}")
@CacheWithTTL.decorate(ttl=60 * 60)
async def get_user_followers(user_id: int, next_cursor: Optional[str] = None):
    """Return a paginated list of the given user's followers.

    Parameters
    ----------
    user_id: string
        The user's ID.
    next_cursor: string | None
        The next search cursor when paginating.

    Returns
    -------
    dict
        Response data
    """
    client = await ClientProxy.get()
    followers, next = await client.get_user_followers(user_id, next_cursor=next_cursor)

    return {
        "next": next,
        "users": [u.to_dict() for u in followers],
    }


@app.get("/following/{user_id}")
@CacheWithTTL.decorate(ttl=60 * 60)
async def get_user_following(user_id: int, next_cursor: Optional[str] = None):
    """Return a paginated list of the people followed by the given user.

    Parameters
    ----------
    user_id: string
        The user's ID.
    next_cursor: string | None
        The next search cursor when paginating.

    Returns
    -------
    dict
        Response data
    """
    client = await ClientProxy.get()
    following, next = await client.get_user_following(user_id, next_cursor=next_cursor)

    return {
        "next": next,
        "users": [u.to_dict() for u in following],
    }


def auth_exception_handler(_: Request, err: Exception) -> JSONResponse:
    """Error handler for AUTH_EXCEPTIONS.
    Reset the client to force a login before enqueuing the exception - it will be later sent to the notifications
    channel.

    See https://fastapi.tiangolo.com/tutorial/handling-errors/.

    Parameters
    ----------
    _: Any
        Not used.
    err: Exception
        The error.

    Returns
    -------
    JSONResponse
        JSON response with the error message.
    """
    ClientProxy.reset()
    enqueue_exception(err)

    return JSONResponse(
        {
            "error": repr(err),
        },
        status_code=500,
    )


def notify_exception_handler(_: Request, err: Exception) -> JSONResponse:
    """Error handler for NOTIFY_EXCEPTIONS.
    Enqueuing the exception so it can be sent to the notifications channel.

    See https://fastapi.tiangolo.com/tutorial/handling-errors/.

    Parameters
    ----------
    _: Any
        Not used.
    err: Exception
        The error.

    Returns
    -------
    JSONResponse
        JSON response with the error message.
    """
    enqueue_exception(err)

    return JSONResponse(
        {
            "error": repr(err),
        },
        status_code=500,
    )


for exc_class in AUTH_EXCEPTIONS:
    app.add_exception_handler(exc_class, auth_exception_handler)

for exc_class in NOTIFY_EXCEPTIONS:
    app.add_exception_handler(exc_class, notify_exception_handler)
