from __future__ import annotations

from datetime import datetime, timezone
import json
from typing import Any, Mapping
import unittest

from bots.python.sdk import connectors


class QueueMarketDataClientTests(unittest.IsolatedAsyncioTestCase):
    async def test_fetch_range_returns_history(self) -> None:
        client = connectors.QueueMarketDataClient()
        await client.put({"symbol": "VN30", "price": 1})
        await client.put({"symbol": "VN30", "price": 2})

        first = await client.fetch()
        second = await client.fetch()
        self.assertEqual(first["price"], 1)
        self.assertEqual(second["price"], 2)

        history = await client.fetch_range()
        self.assertEqual(len(history), 2)
        self.assertEqual(history[0]["price"], 1)


class FakeRedis:
    def __init__(self, responses: list[list[object]]) -> None:
        self._responses = responses
        self.commands: list[tuple[object, ...]] = []

    async def execute_command(self, *args):  # type: ignore[override]
        self.commands.append(args)
        if not self._responses:
            return None
        return self._responses.pop(0)


class FakeRedisStream:
    def __init__(self) -> None:
        self.entries: list[list] = []
        self.calls: list[tuple[dict, int, Optional[int]]] = []

    def push(self, entry_id: str, payload: Mapping[str, Any], stream: str = "ssi_ps_stream:ABC") -> None:
        self.entries.append([stream, [(entry_id, payload)]])

    async def xread(self, streams, count: int = 1, block: Optional[int] = None):  # type: ignore[override]
        self.calls.append((streams, count, block))
        if not self.entries:
            return []
        return [self.entries.pop(0)]


class RedisTimeSeriesMarketDataClientTests(unittest.IsolatedAsyncioTestCase):
    async def test_fetch_returns_latest_sample(self) -> None:
        redis = FakeRedis([(1700, b"100.5"), (1700, b"250")])
        client = connectors.RedisTimeSeriesMarketDataClient(
            redis,
            "markets:vn30:price",
            symbol="VN30",
            extra_fields={"volume": "markets:vn30:volume"},
        )

        sample = await client.fetch()

        self.assertEqual(sample["price"], 100.5)
        self.assertEqual(sample["volume"], 250.0)
        self.assertEqual(sample["symbol"], "VN30")
        self.assertIn("timestamp", sample)
        self.assertEqual(redis.commands[0], ("TS.GET", "markets:vn30:price"))
        self.assertEqual(redis.commands[1], ("TS.GET", "markets:vn30:volume"))

    async def test_fetch_range_converts_timestamps(self) -> None:
        start = datetime(2024, 1, 1, tzinfo=timezone.utc)
        end = start.replace(minute=start.minute + 1)
        redis = FakeRedis([[[1700, "100.5"], [1701, "101.5"]]])
        client = connectors.RedisTimeSeriesMarketDataClient(
            redis,
            "markets:vn30:price",
            symbol="VN30",
        )
        samples = await client.fetch_range(start=start, end=end, count=100)
        self.assertEqual(len(samples), 2)
        self.assertEqual(samples[0]["price"], 100.5)
        self.assertEqual(redis.commands[0][0], "TS.RANGE")
        self.assertEqual(redis.commands[0][1], "markets:vn30:price")
        # Ensure timestamp conversion
        self.assertTrue(all(isinstance(item["timestamp"], datetime) for item in samples))


class RedisStreamMarketDataClientTests(unittest.IsolatedAsyncioTestCase):
    async def test_fetch_decodes_json_payload(self) -> None:
        stream = FakeRedisStream()
        payload = json.dumps({"code": "ABC", "price": 1.23, "timestamp": "2024-01-01T00:00:00Z"})
        stream.push("1700-0", {"payload": payload})

        client = connectors.RedisStreamMarketDataClient(stream, "ssi_ps_stream:ABC", last_id="0-0", block_ms=50)
        snapshot = await client.fetch()

        self.assertEqual(snapshot["code"], "ABC")
        self.assertEqual(snapshot["price"], 1.23)
        self.assertIsInstance(snapshot["timestamp"], datetime)
        self.assertEqual(snapshot["stream_id"], "1700-0")

    async def test_fetch_returns_empty_when_no_data(self) -> None:
        stream = FakeRedisStream()
        client = connectors.RedisStreamMarketDataClient(stream, "ssi_ps_stream:ABC", block_ms=1)
        snapshot = await client.fetch()
        self.assertEqual(snapshot, {})

if __name__ == "__main__":  # pragma: no cover
    unittest.main()
