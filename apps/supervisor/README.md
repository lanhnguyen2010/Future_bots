# Supervisor Service

Golang control-plane service exposing REST and WebSocket APIs for managing bot lifecycles. Implements desired state reconciliation against Kubernetes and publishes runtime commands to Redis/Kafka.

## Database Migrations

Set `SUPERVISOR_DATABASE_URL` (and optionally `SUPERVISOR_DATABASE_DRIVER`, default `pgx`) to apply the bundled migrations on startup. The SQL lives under `internal/migrations/sql` and follows golang-migrate conventions, so the same files can be executed with the CLI if you prefer manual control. When the variable is unset migrations are skipped, allowing stateless development flows.

## Redis Telemetry

Provide `SUPERVISOR_REDIS_ADDR` (and optional `SUPERVISOR_REDIS_USERNAME`,
`SUPERVISOR_REDIS_PASSWORD`) to enable RedisTimeSeries telemetry for bot desired
state changes. When configured, the service records configuration revisions and
enable/disable toggles under `bots:<bot_id>:config_rev|enabled`. Tweak the
retention window with `SUPERVISOR_REDIS_METRIC_RETENTION` (default 30 days).
