from __future__ import annotations

import logging
from dataclasses import dataclass, field
from typing import Any, Mapping, MutableMapping, Optional

from . import connectors


@dataclass
class BotContext:
    """
    Wiring for platform integrations available to bots at runtime.

    Attributes:
        bot_id: Unique identifier of the bot.
        account_id: Account the bot operates under.
        market_data: Provider returning the latest market snapshot.
        orders: Publisher for order intents emitted by the bot.
        state: Key-value store used for bot local state.
        control: Optional control channel for out-of-band commands.
        heartbeat: Optional heartbeat sink for liveness updates.
        logger: Python logger preconfigured with contextual metadata.
        extras: Mutable mapping for attaching additional integrations.
    """

    bot_id: str
    account_id: str
    market_data: connectors.MarketDataClient = field(
        default_factory=connectors.NullMarketDataClient
    )
    orders: connectors.OrderPublisher = field(
        default_factory=connectors.StdoutOrderPublisher
    )
    state: connectors.StateStore = field(default_factory=connectors.InMemoryStateStore)
    control: Optional[connectors.ControlChannel] = field(
        default_factory=connectors.NullControlChannel
    )
    heartbeat: Optional[connectors.HeartbeatPublisher] = field(
        default_factory=connectors.NullHeartbeatPublisher
    )
    logger: logging.Logger = field(default_factory=lambda: logging.getLogger("bot"))
    extras: MutableMapping[str, Any] = field(default_factory=dict)

    def with_logger(self, logger: logging.Logger) -> "BotContext":
        """
        Returns a shallow copy of the context with the provided logger.
        Useful for injecting structured logging stacks.
        """
        clone = dataclass_replace(self)
        clone.logger = logger
        return clone

    def add_extra(self, key: str, value: Any) -> None:
        """Attach arbitrary extra context information."""
        self.extras[key] = value

    def get_extra(self, key: str, default: Any = None) -> Any:
        """Fetch a previously stored extra value."""
        return self.extras.get(key, default)


def dataclass_replace(ctx: BotContext) -> BotContext:
    """Helper to clone dataclass while keeping field defaults safe."""
    return BotContext(
        bot_id=ctx.bot_id,
        account_id=ctx.account_id,
        market_data=ctx.market_data,
        orders=ctx.orders,
        state=ctx.state,
        control=ctx.control,
        heartbeat=ctx.heartbeat,
        logger=ctx.logger,
        extras=dict(ctx.extras),
    )
