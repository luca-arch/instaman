"""Slow in-memory caching system.
"""

import time
from functools import wraps
from typing import Any, Callable, Dict, Optional, Tuple

# Sentinel for separating args from kwargs.
# This is similar to what functools.lru_cache() does.
mark = hash(object())


class CacheWithTTL:
    """In-memory cache with TTL.

    Weakly implemented, this makes an hash out of the received args and kwargs and then performs
    an O(1) lookup inside a dictionary.
    """

    entries: Dict[str, Any]
    expiry: Dict[str, float]

    def __init__(self) -> None:
        self.entries = {}
        self.expiry = {}

    def evict(self) -> None:
        """Evict all expired items from the cache."""
        now = time.time()
        expired = (k for k, t in self.expiry.items() if 0 < t < now)

        for k in expired:
            del self.entries[k]
            del self.expiry[k]

    def get(self, key: str) -> Tuple[Any, bool]:
        """Lookup item in the cache.

        Parameters
        ----------
        key: str
            The item's unique key.

        Returns
        -------
        [Any, bool]
            A tuple containing the item, and whether the item was found.
        """
        if key not in self.entries:
            return None, False

        if 0 < self.expiry[key] < time.time():
            del self.entries[key]
            del self.expiry[key]

            return None, False

        return self.entries[key], True

    def set(self, key: str, value: Any, ttl: Optional[int] = None) -> None:
        """Store a new item in the cache.

        Parameters
        ----------
        key: str
            The item's unique key.

        value: Any
            The value to store,

        ttl: int | None
            The item's TTL after which it's removed from the cache.
        """
        self.evict()
        self.entries[key] = value
        self.expiry[key] = (time.time() + ttl) if ttl else 0

    @classmethod
    def decorate(cls, ttl: Optional[int] = None):
        """Function decorator."""
        mem = cls()

        def wrapper(coro: Callable):
            @wraps(coro)
            async def wrapper_func(*args, **kwargs):
                key = str(args) + str(mark) + str(tuple(sorted(kwargs.items())))

                value, found = mem.get(key)
                if found:
                    return value

                value = await coro(*args, **kwargs)
                mem.set(key, value, ttl=ttl)

                return value

            return wrapper_func

        return wrapper
