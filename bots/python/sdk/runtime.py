from __future__ import annotations

import asyncio
import logging
import signal
from dataclasses import dataclass, field
from time import monotonic
from typing import Any, Iterable, Mapping, Optional

from .base import BotBase
from .context import BotContext
from .connectors import ensure_async


@dataclass
class RuntimeConfig:
    """User-tunable configuration governing the bot runtime loop."""

    poll_interval: float = 1.0
    heartbeat_interval: float = 5.0
    max_consecutive_errors: int = 5
    graceful_shutdown_timeout: float = 5.0


@dataclass
class RuntimeStats:
    """Aggregate statistics emitted by the runtime."""

    ticks: int = 0
    orders_published: int = 0
    consecutive_errors: int = 0
    last_error: Optional[str] = None
    last_heartbeat_at: Optional[float] = None


class BotRuntime:
    """
    Coordinates the lifecycle of a `BotBase` implementation.  Handles
    polling, order publication, control messages, and graceful shutdown.
    """

    def __init__(
        self,
        bot: BotBase,
        ctx: BotContext,
        config: RuntimeConfig | None = None,
        loop: asyncio.AbstractEventLoop | None = None,
    ) -> None:
        self.bot = bot
        self.ctx = ctx
        self.config = config or RuntimeConfig()
        self.loop = loop or asyncio.get_event_loop()
        self.stats = RuntimeStats()
        self._stop_event = asyncio.Event()
        self._stop_reason = "runtime_shutdown"
        self._logger = ctx.logger.getChild("runtime")
        self._shutdown_future: Optional[asyncio.Future[None]] = None
        self._signal_handlers_installed = False

    async def run_forever(self) -> None:
        """
        Main entry point that drives the bot until a stop is requested.
        Suitable for `asyncio.run(BotRuntime(...).run_forever())`.
        """
        self._install_signal_handlers()
        self._logger.info("bot_runtime_starting")
        await ensure_async(self.bot.on_start())
        try:
            await self._main_loop()
        except asyncio.CancelledError:  # pragma: no cover - defensive
            self._logger.info("bot_runtime_cancelled")
        finally:
            await self._shutdown()

    async def stop(self, reason: str = "external-request") -> None:
        """Triggers a graceful stop of the runtime."""
        if not self._stop_event.is_set():
            self._stop_reason = reason
            self._stop_event.set()
        await self._wait_for_shutdown()

    async def _main_loop(self) -> None:
        """Executes the bot polling loop until stop conditions are met."""
        poll_interval = max(self.config.poll_interval, 0.05)
        heartbeat_interval = max(self.config.heartbeat_interval, 0.1)
        next_heartbeat = monotonic()

        while not self._stop_event.is_set():
            tick_started = monotonic()

            if self.ctx.control:
                await self._drain_control_messages()
                if self._stop_event.is_set():
                    break

            try:
                market = await self.ctx.market_data.fetch()
                intents = await ensure_async(self.bot.on_tick(market))
                intents_iter = intents or []
                published = await self._publish_intents(intents_iter)
                self.stats.orders_published += published
                self.stats.consecutive_errors = 0
                self.stats.ticks += 1
            except Exception as exc:  # pylint: disable=broad-except
                self.stats.consecutive_errors += 1
                self.stats.last_error = repr(exc)
                self._logger.exception("bot_tick_failed")
                if (
                    self.config.max_consecutive_errors
                    and self.stats.consecutive_errors >= self.config.max_consecutive_errors
                ):
                    self._logger.error(
                        "bot_max_consecutive_errors_reached",
                        extra={"count": self.stats.consecutive_errors},
                    )
                    await self.stop("max-errors")
                    break

            now = monotonic()
            if self.ctx.heartbeat and now >= next_heartbeat:
                await self._emit_heartbeat()
                next_heartbeat = now + heartbeat_interval

            elapsed = monotonic() - tick_started
            sleep_for = max(poll_interval - elapsed, 0)
            try:
                await asyncio.wait_for(self._stop_event.wait(), timeout=sleep_for)
            except asyncio.TimeoutError:
                continue

    async def _publish_intents(self, intents: Iterable[Mapping[str, Any]]) -> int:
        count = 0
        for intent in intents:
            if intent is None:
                continue
            await self.ctx.orders.publish(intent)
            count += 1
        return count

    async def _drain_control_messages(self) -> None:
        assert self.ctx.control is not None
        while True:
            message = await self.ctx.control.receive(timeout=0)
            if not message:
                break
            await ensure_async(self.bot.on_event(message))
            msg_type = str(message.get("type", "")).lower()
            if msg_type in {"bot.stop", "stop"}:
                reason = message.get("reason", "stop-command")
                await self.stop(str(reason))
                break

    async def _emit_heartbeat(self) -> None:
        payload = {
            "bot_id": self.ctx.bot_id,
            "account_id": self.ctx.account_id,
            "status": await ensure_async(self.bot.health()),
            "stats": {
                "ticks": self.stats.ticks,
                "orders_published": self.stats.orders_published,
                "last_error": self.stats.last_error,
            },
        }
        self.stats.last_heartbeat_at = monotonic()
        if self.ctx.heartbeat:
            await self.ctx.heartbeat.publish(payload)

    async def _shutdown(self) -> None:
        try:
            await ensure_async(self.bot.on_stop(self._stop_reason))
        except Exception:  # pragma: no cover - best effort
            self._logger.exception("bot_on_stop_failed")
        self._logger.info("bot_runtime_stopped", extra={"reason": self._stop_reason})

    def _install_signal_handlers(self) -> None:
        if self._signal_handlers_installed:
            return
        loop = asyncio.get_event_loop()
        for sig in (signal.SIGTERM, signal.SIGINT):
            try:
                loop.add_signal_handler(sig, lambda s=sig: asyncio.create_task(self.stop(f"signal-{s.name.lower()}")))
            except (NotImplementedError, RuntimeError):  # pragma: no cover - e.g., Windows or non-main thread
                continue
        self._signal_handlers_installed = True

    async def _wait_for_shutdown(self) -> None:
        if self._shutdown_future is None:
            self._shutdown_future = asyncio.create_task(self._stop_event.wait())
        await self._shutdown_future
