# Future Bots Monorepo

This repository houses the full system design and scaffolding for the Qubit Bot Trading Platform targeting HOSE futures. The codebase is organized as a polyglot monorepo so cross-cutting changes to bots, services, and infrastructure can be developed together.

## Repository Layout

```
.
├── apps
│   ├── executor/           # Golang trade executor microservice
│   ├── reports/            # Golang reporting & analytics service
│   ├── risk/               # Golang risk engine service
│   └── supervisor/         # Golang control plane APIs & websocket hub
├── bots
│   └── python
│       ├── sdk/            # Shared bot SDK contract + runner libraries
│       └── samples/        # Example strategy packages for validation
├── docs/                   # Architecture & decision records
├── proto/                  # Kafka protobuf schemas shared across services
├── infra
│   ├── k8s/                # Kubernetes manifests & helm charts
│   └── sql/                # TimescaleDB schema, migrations, seed data
├── scripts/                # Tooling, CI helpers, dev environment scripts
└── Makefile                # Developer convenience commands
```

## Getting Started

1. Review the full platform design in [`docs/qubit-bot-platform.md`](docs/qubit-bot-platform.md).
2. Use the provided `Makefile` targets to bootstrap local dependencies.
3. Implement services and bots inside the respective directories, following the design contract.
4. For Go services, the repository ships with a [`go.work`](go.work) workspace and individual modules under `apps/`. Use
   `go run ./cmd/<service>` from each module (or the Makefile targets) to execute the scaffolded HTTP APIs.

## Makefile Targets

Run `make help` to see available commands. Key targets include:

- `make up` – start TimescaleDB and Redis via Docker Compose.
- `make seed` – apply core schemas to the Timescale database.
- `make api` – run the Supervisor/Executor Go services locally.
- `make bot` – execute a Python bot using the SDK runner.
- `make down` – stop all local dependencies and clean volumes.

## Protobuf Schemas

Kafka topic payloads are defined under [`proto/`](proto/). The schemas capture
order intents, execution events, risk alerts, and supervisor bot commands. Use
`protoc` (the dev container ships with the compiler) to generate language
bindings for services and bots. See [`proto/README.md`](proto/README.md) for
details and sample commands.

## Dev Container

The repository includes a VS Code [dev container](.devcontainer) preloaded with
Go, Python, Node.js, pnpm, Poetry, `protoc`, and `buf`. Open the folder in VS
Code and run "Reopen in Container" to provision a reproducible development
environment with the required toolchain and sensible editor defaults.

Additional details are documented in the architecture guide.

## Database Migrations

Each Go service ships with lightweight SQL migrations organised under `internal/migrations/sql` using the same naming convention as [golang-migrate](https://github.com/golang-migrate/migrate) (`0001_description.up.sql` / `0001_description.down.sql`). The shared `libs/go/platform/db` helper loads these embedded files at startup. Set the corresponding `*_DATABASE_URL` environment variable (e.g. `SUPERVISOR_DATABASE_URL`) before starting a service to apply migrations automatically. The optional `*_DATABASE_DRIVER` variable defaults to `pgx`, allowing teams to point at TimescaleDB or any Postgres-compatible database. The migrations remain CLI-compatible, so teams can also run `golang-migrate` manually against the same files when operating outside the services.
