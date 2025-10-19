from __future__ import annotations

import asyncio
import logging
import unittest

from bots.python.sdk import BotBase, BotContext, BotRuntime, RuntimeConfig, connectors


class SampleBot(BotBase):
    def __init__(self, config, secrets, ctx):
        super().__init__(config, secrets, ctx)
        self.tick_count = 0
        self.stop_reason = None

    def on_tick(self, market):
        self.tick_count += 1
        return [{"symbol": market.get("symbol", "SYM"), "side": "buy", "qty": self.tick_count}]

    def on_stop(self, reason: str) -> None:
        self.stop_reason = reason


class ExplodingBot(BotBase):
    def on_tick(self, market):
        raise RuntimeError("boom")


class BotRuntimeTests(unittest.IsolatedAsyncioTestCase):
    async def test_runtime_publishes_orders_and_stops_on_command(self):
        market = connectors.QueueMarketDataClient()
        orders = connectors.ListOrderPublisher()
        control = connectors.QueueControlChannel()
        heartbeat = connectors.HeartbeatBuffer()
        ctx = BotContext(
            bot_id="bot-1",
            account_id="acct-1",
            market_data=market,
            orders=orders,
            control=control,
            heartbeat=heartbeat,
            logger=logging.getLogger("test.bot"),
        )
        bot = SampleBot({}, {}, ctx)
        runtime = BotRuntime(
            bot,
            ctx,
            RuntimeConfig(poll_interval=0.05, heartbeat_interval=0.05),
        )

        async def feeder():
            await market.put({"symbol": "VN30", "price": 1})
            await market.put({"symbol": "VN30", "price": 2})
            await asyncio.sleep(0.05)
            await control.put({"type": "bot.stop", "reason": "test"})

        await asyncio.gather(runtime.run_forever(), feeder())

        self.assertGreaterEqual(bot.tick_count, 2)
        self.assertEqual(bot.stop_reason, "test")
        self.assertGreaterEqual(len(orders.items), 2)
        self.assertTrue(heartbeat.items)

    async def test_runtime_stops_after_consecutive_errors(self):
        market = connectors.StaticMarketDataClient({"symbol": "VN30"})
        orders = connectors.ListOrderPublisher()
        ctx = BotContext(
            bot_id="bot-err",
            account_id="acct-err",
            market_data=market,
            orders=orders,
            control=connectors.NullControlChannel(),
            heartbeat=connectors.NullHeartbeatPublisher(),
            logger=logging.getLogger("test.errbot"),
        )
        bot = ExplodingBot({}, {}, ctx)
        runtime = BotRuntime(
            bot,
            ctx,
            RuntimeConfig(poll_interval=0.01, max_consecutive_errors=3),
        )

        await runtime.run_forever()

        self.assertEqual(runtime.stats.consecutive_errors, 3)
        self.assertEqual(bot.ctx.state, ctx.state)  # ensure context intact


if __name__ == "__main__":  # pragma: no cover
    unittest.main()
