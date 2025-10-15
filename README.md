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

- `make up` – start TimescaleDB, Redis, Kafka, and Kafka UI via Docker Compose.
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

## Local Runtime Options

### Docker Compose

The root [`docker-compose.yml`](docker-compose.yml) provisions the platform
dependencies—TimescaleDB, Redis, Kafka/ZooKeeper, and the Kafka UI—along with a
`dev-env` container that mirrors the repository's devcontainer image. Use the
Makefile target above or run `docker compose up -d` manually to bring the stack
online. Default credentials match the documentation: `qubit`/`qubit` for the
database and plaintext listeners on `localhost:9092` and `localhost:9094` for
Kafka.

### Minikube Deployment

For a lightweight Kubernetes install, execute
[`scripts/deploy_minikube.sh`](scripts/deploy_minikube.sh). The script starts (or
reuses) a `qubit-bots` Minikube profile, ensures the `trading` namespace exists,
and applies the manifests in [`infra/k8s`](infra/k8s). Override CPU, memory, or
namespace defaults by exporting `MINIKUBE_CPUS`, `MINIKUBE_MEMORY`, or
`K8S_NAMESPACE` before running the script.

## Dev Container

The repository includes a VS Code [dev container](.devcontainer) preloaded with
Go, Python, Node.js, pnpm, Poetry, `protoc`, and `buf`. Open the folder in VS
Code and run "Reopen in Container" to provision a reproducible development
environment with the required toolchain and sensible editor defaults.

Additional details are documented in the architecture guide.

## Database Migrations

Each Go service ships with lightweight SQL migrations organised under `internal/migrations/sql` using the same naming convention as [golang-migrate](https://github.com/golang-migrate/migrate) (`0001_description.up.sql` / `0001_description.down.sql`). The shared `libs/go/platform/db` helper loads these embedded files at startup. Set the corresponding `*_DATABASE_URL` environment variable (e.g. `SUPERVISOR_DATABASE_URL`) before starting a service to apply migrations automatically. The optional `*_DATABASE_DRIVER` variable defaults to `pgx`, allowing teams to point at TimescaleDB or any Postgres-compatible database. The migrations remain CLI-compatible, so teams can also run `golang-migrate` manually against the same files when operating outside the services.

## API Documentation

The Supervisor, Executor, Risk, and Reports services now expose OpenAPI 3.0 specifications along with an embedded Swagger UI. When running any service locally, visit `http://localhost:<port>/docs` to browse interactive documentation, or fetch the machine-readable contract at `/openapi.json` for client generation.
