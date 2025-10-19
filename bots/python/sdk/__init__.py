"""
Python Bot SDK for the Qubit trading platform.

The SDK exposes a lightweight runtime that bot authors can embed inside
their containers.  It provides:

- `BotBase` – abstract lifecycle contract that user bots implement.
- `BotContext` – wiring for platform integrations (market data, orders,
  state, control channel, heartbeat publisher).
- `BotRuntime` – orchestrates the lifecycle, polling market data,
  invoking `on_tick`, publishing order intents, and emitting heartbeats.
- Connector protocols and simple in-memory defaults for local testing.
"""

from .base import BotBase
from .context import BotContext
from .runtime import BotRuntime, RuntimeConfig, RuntimeStats
from . import connectors

__all__ = [
    "BotBase",
    "BotContext",
    "BotRuntime",
    "RuntimeConfig",
    "RuntimeStats",
    "connectors",
]
