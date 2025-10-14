# Reports Service

Golang service that aggregates historical orders, executions, and PnL from TimescaleDB to produce dashboards, scheduled reports, and ad-hoc analytics consumed by the web dashboard.

## Database Migrations

Set `REPORTS_DATABASE_URL` (and optionally `REPORTS_DATABASE_DRIVER`, default `pgx`) to bootstrap job and snapshot tables before serving requests. Migrations are stored in `internal/migrations/sql` using golang-migrate-compatible filenames so they can also be applied with the CLI.
