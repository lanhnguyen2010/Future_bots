from __future__ import annotations

from abc import ABC, abstractmethod
from typing import Any, Dict, Iterable, Mapping, Optional


class BotBase(ABC):
    """
    Base lifecycle contract for trading bots.

    Bot authors subclass `BotBase` and implement the lifecycle hooks
    described below.  The runtime will instantiate the bot with config,
    secrets, and a `BotContext` instance that exposes platform services.
    """

    def __init__(self, config: Mapping[str, Any], secrets: Mapping[str, Any], ctx: "BotContext") -> None:  # pragma: no cover - BotContext is injected
        self.config = dict(config)
        self.secrets = dict(secrets)
        self.ctx = ctx

    def on_start(self) -> None:
        """
        Invoked once after the bot is constructed and before the main loop.
        Subclasses can override this method to perform warm-up work, load
        state, or validate configuration.  Return value is ignored.
        """

    @abstractmethod
    def on_tick(self, market: Mapping[str, Any]) -> Optional[Iterable[Mapping[str, Any]]]:
        """
        Invoked on every market data poll.  Implementations may return an
        iterable of order intent payloads which will be forwarded to the
        executor service.  Returning `None` or an empty iterable results
        in no orders being published.
        """

    def on_event(self, event: Mapping[str, Any]) -> None:
        """
        Hook for control/event messages (e.g., STOP commands, fills).
        Override to react to asynchronous platform events.
        """

    def on_stop(self, reason: str) -> None:
        """
        Invoked exactly once when the runtime is shutting down.  A `reason`
        string is provided for diagnostics (e.g., "stop-command" or
        "max-errors").  Implementations should attempt to flush state but
        must return quickly to avoid delaying shutdown.
        """

    def health(self) -> Dict[str, Any]:
        """
        Returns a serialisable representation of bot health.  Defaults to
        `{ "ok": True }` but can be overridden to expose richer status
        indicators (e.g., last tick latency, inventory levels).
        """
        return {"ok": True}
