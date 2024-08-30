"""Telegram notifications for errors

This module allows other modules to push Exception instances into different
queues so that errors can be notified to the program owner via Telegram.

Care should be taken when using the `enqueue_exception()` funcction since
Exceptions are grouped by class name, and if there are multiple exceptions of
the same class, only the stack trace from the first exception will be sent
along the notification.
"""

import json
import os
import traceback
from typing import DefaultDict, List
import aiohttp
import asyncio
import logging
from collections import defaultdict

# Telegram BOT token.
BOT_TOKEN = os.environ["TG_BOT_TOKEN"]

# Telegram channel where to notify errors.
CHANNEL_ID = os.environ["TG_CHANNEL"]

# Error queues (keys are the exceptions' class names)
QUEUES: DefaultDict[str, asyncio.Queue[Exception]] = defaultdict(asyncio.Queue)


def enqueue_exception(err: Exception) -> None:
    """
    Enqueue the received error, grouped by its classname.
    If there are more than 5000 errors in the queue, this function will do nothing but log a warning.

    Parameters
    ----------
    err: Exception
        The exception to enqueue for sending.
    """
    qname = err.__class__.__name__
    queue = QUEUES[qname]
    qsize = queue.qsize()

    if qsize > 5000:
        logging.warning("Exception queue %s already contains %d errors", qname, qsize)
        return

    queue.put_nowait(err)


def get_message_text(errors: List[Exception]) -> str:
    error, errors_count = errors[0], len(errors)
    text = [
        f"💥 **Instaproxy error ({type(error).__name__}): {error}**",
        "",
        "```",
        "".join(traceback.TracebackException.from_exception(error).format()),
        "```",
    ]

    if errors_count > 1:
        text += [
            "",
            f"This error occurred {errors_count} times, only the first stack trace is attached",
        ]

    return "\n".join(text)


async def notify_exception_group(queue: asyncio.Queue[Exception]) -> bool:
    """
    Notify a list of exceptions via Telegram.
    https://core.telegram.org/bots/api#sendmessage

    Parameters
    ----------
    queue: asyncio.Queue
        The queue to drain and send the notification for.

    Returns
    -------
    bool
        Whether a message was sent.
    """
    errors: List[Exception] = []

    # Drain queue first
    while True:
        try:
            errors.append(queue.get_nowait())
        except asyncio.QueueEmpty:
            break

    if not errors:
        logging.warning(
            "notify_exception_group was invoked on an empty queue, skipping..."
        )

        return False

    data = {
        "chat_id": CHANNEL_ID,
        "link_preview_options": json.dumps({"is_disabled": True}),
        "parse_mode": "MarkdownV2",
        "text": get_message_text(errors),
    }

    try:
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"https://api.telegram.org/bot{BOT_TOKEN}/sendMessage", data=data
            ) as response:
                if response.status == 200:
                    return True
    except aiohttp.ClientError as client_exception:
        logging.warning("Failed to post Telegram message.")
        logging.exception(client_exception)

    try:
        resp = await response.json()
        desc = resp["descriptions"].values()
    except:  # noqa: E722
        logging.warning("Failed to post Telegram message. Status: %s", response.status)
    else:
        logging.debug("Telegram response: %s", resp)
        logging.warning("Telegram error (%s): %s", response.status, "; ".join(desc))

    if response.status >= 500:
        # Re-queue for retry.
        await queue.put(errors[0])
        await asyncio.sleep(30)

    return False


async def start_notifier() -> asyncio.Task:
    """
    Start `watch_queues()` in a backgroung task and return such task.

    Returns
    -------
    asyncio.Task
        the active background task.
    """
    task = asyncio.create_task(watch_queues())

    return task


async def watch_queues() -> None:
    """
    Keep polling all queues and notify any error to the Telegram channel.
    This coroutine:
    - will sleep for 30 seconds if there are no queues in the dictionary
    - will sleep 10 seconds between each message
    """
    while True:
        if not QUEUES:
            await asyncio.sleep(30)

            continue

        for key in list(QUEUES.keys()):
            queue = QUEUES[key]

            if queue.empty():
                del QUEUES[key]
                continue

            sent = await notify_exception_group(queue)
            if sent:
                # Do not evict the queue from the dictionary here, new messages might have appeared while sending!
                await asyncio.sleep(10)
