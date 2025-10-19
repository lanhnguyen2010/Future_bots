# Python Bot SDK

Defines the abstract `BotBase` contract and runtime helpers for connecting to market data, Redis state, and Kafka order buses. Bot authors subclass `BotBase` and implement lifecycle hooks (`on_start`, `on_tick`, etc.).

## Modules

- `bot_base.py` – lifecycle contract bots must implement.
- `context.py` – typed wiring for platform integrations (market data feed, order publisher, state store, control channel, heartbeat).
- `connectors.py` – connector protocols and lightweight in-memory implementations for local development.
- `connectors.py` now also ships `RedisTimeSeriesMarketDataClient` and
  `RedisStreamMarketDataClient` for consuming historical snapshots and live
  ticks directly from Redis.
- `connectors.py` now also ships a `RedisTimeSeriesMarketDataClient` for bots that
  consume ticker data stored in RedisTimeSeries. Install the optional dependency
  with `pip install redis` when using it.
- `runtime.py` – asynchronous runtime driving the bot lifecycle with polling, order publication, control handling and heartbeats.
- `tests/` – unit tests demonstrating how to wire the runtime using the in-memory connectors.

## Quickstart

```python
import asyncio
from bots.python.sdk import BotBase, BotContext, BotRuntime, RuntimeConfig, connectors


class ExampleBot(BotBase):
    def on_start(self):
        self.ctx.logger.info("ready")

    def on_tick(self, market):
        price = market.get("price")
        if not price:
            return []
        return [{"symbol": market["symbol"], "side": "buy", "qty": 1, "price": price}]


async def main():
    ctx = BotContext(
        bot_id="bot-1",
        account_id="acct-1",
        market_data=connectors.QueueMarketDataClient(),
        orders=connectors.StdoutOrderPublisher(),
    )
    bot = ExampleBot({}, {}, ctx)
    runtime = BotRuntime(bot, ctx, RuntimeConfig(poll_interval=0.2))
    await runtime.run_forever()


if __name__ == "__main__":
    asyncio.run(main())

# Using RedisTimeSeries in a bot

```python
from datetime import datetime, timedelta, timezone
from redis.asyncio import Redis

from bots.python.sdk import connectors


redis_client = Redis(host="localhost", port=6379)
market_data = connectors.RedisTimeSeriesMarketDataClient(
    redis_client,
    "markets:vn30f1m:price",
    symbol="VN30F1M",
    extra_fields={"volume": "markets:vn30f1m:volume"},
)

latest = await market_data.fetch()
range_samples = await market_data.fetch_range(
    start=datetime.now(timezone.utc) - timedelta(minutes=30)
)

# Listening to Redis Streams

stream_client = connectors.RedisStreamMarketDataClient(
    redis_client,
    "ssi_ps_stream:VN30F1M",
    last_id="0-0",
    block_ms=1000,
)

tick = await stream_client.fetch()  # blocks until a new snapshot arrives
```
```
