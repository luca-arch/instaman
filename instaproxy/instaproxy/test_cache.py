import pytest
import time
from .cache import CacheWithTTL


@pytest.mark.anyio
async def test_cache():
    cache, item = CacheWithTTL(), object()

    cache.set("my-key", item)

    assert "my-key" in cache.entries
    assert "my-key" in cache.expiry

    assert cache.entries["my-key"] == item
    assert cache.expiry["my-key"] == 0

    assert cache.get("my-key") == (item, True)


@pytest.mark.anyio
async def test_with_ttl():
    delta = 0.1

    cache, item = CacheWithTTL(), object()
    now, ttl = time.time(), 100

    cache.set("my-key", item, ttl=ttl)
    assert (now + ttl - delta) < cache.expiry["my-key"] < (now + ttl + delta)
    assert cache.get("my-key") == (item, True)


@pytest.mark.anyio
async def test_evict():
    cache, item = CacheWithTTL(), object()

    cache.set("my-key", item, ttl=100)
    assert cache.get("my-key") == (item, True)

    # Force expiry then evict
    cache.expiry["my-key"] = 1
    cache.evict()

    assert "my-key" not in cache.entries
    assert "my-key" not in cache.expiry
    assert cache.get("my-key") == (None, False)


@pytest.mark.anyio
async def test_evict_on_update():
    cache, item1, item2 = CacheWithTTL(), object(), object()

    cache.set("item-1", item1, ttl=100)
    assert cache.get("item-1") == (item1, True)

    # Force expiry then cache another item
    cache.expiry["item-1"] = 1
    cache.set("item-2", item2, ttl=100)

    assert cache.get("item-1") == (None, False)
    assert cache.get("item-2") == (item2, True)
