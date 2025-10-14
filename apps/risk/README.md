# Risk Engine Service

Golang service responsible for real-time exposure checks, circuit breakers, and account-level guardrails. Provides synchronous decision APIs to the Trade Executor and emits alerts via Kafka.

## Database Migrations

Set `RISK_DATABASE_URL` (and optionally `RISK_DATABASE_DRIVER`, default `pgx`) to bootstrap reference tables for limits and alert history when the service starts. The migrations under `internal/migrations/sql` follow golang-migrate conventions for easy CLI usage outside the service.
