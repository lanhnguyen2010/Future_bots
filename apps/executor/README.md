# Trade Executor Service

Golang microservice that consumes order intents from Kafka, invokes the risk engine, submits orders to broker APIs with idempotency guarantees, and persists order/execution lifecycle events to TimescaleDB.

## Database Migrations

Set `EXECUTOR_DATABASE_URL` (and optionally `EXECUTOR_DATABASE_DRIVER`, default `pgx`) to run the packaged schema migrations automatically on boot. Migrations live in `internal/migrations/sql` following golang-migrate naming, making them compatible with the CLI as well.
