# Python Bot SDK

Defines the abstract `BotBase` contract and runtime helpers for connecting to market data, Redis state, and Kafka order buses. Bot authors subclass `BotBase` and implement lifecycle hooks (`on_start`, `on_tick`, etc.).
