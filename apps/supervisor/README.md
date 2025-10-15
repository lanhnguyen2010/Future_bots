# Supervisor Service

Golang control-plane service exposing REST and WebSocket APIs for managing bot lifecycles. Implements desired state reconciliation against Kubernetes and publishes runtime commands to Redis/Kafka.

## Database Migrations

Set `SUPERVISOR_DATABASE_URL` (and optionally `SUPERVISOR_DATABASE_DRIVER`, default `pgx`) to apply the bundled migrations on startup. The SQL lives under `internal/migrations/sql` and follows golang-migrate conventions, so the same files can be executed with the CLI if you prefer manual control. When the variable is unset migrations are skipped, allowing stateless development flows.
