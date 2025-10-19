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
- `make marketdata-seed` – inject sample `RedisTimeSeries` ticks for quick local testing.
- `make producer` – parse SSI sample data and publish `SsiPsSnapshot` records to Kafka.
- `make consumer` – subscribe to `ssi_ps` Kafka topic and persist snapshots to RedisTimeSeries.
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

Redis is provided via the `redis/redis-stack` image so the `RedisTimeSeries`
module is available out of the box; the compose health check now verifies the
module is loaded before other services attempt to connect.

### RedisTimeSeries Market Data

Sample ticker data can be seeded with `make marketdata-seed`, which executes
`apps/reports/cmd/seed-marketdata` against the local Redis instance. The command
creates price and volume time series under `markets:<ticker>:price|volume` (by
default `VN30F1M`). Adjust the target using environment variables, for example:

```bash
MARKETDATA_TICKER=VN30F2M MARKETDATA_REDIS_ADDR=localhost:6379 make marketdata-seed
```

Bots can consume the data using the new
`RedisTimeSeriesMarketDataClient` available from the Python SDK
(`bots/python/sdk/connectors.py`), which exposes `fetch` and `fetch_range`
helpers for retrieving the latest point or a bounded time window.

### Kafka Topics

The stack autocrates Kafka topics, but you can explicitly create the Hose
PowerScreen channel used by the stock parser via:

```bash
kafka-topics --bootstrap-server localhost:9092 --create --topic ssi_ps --partitions 6 --replication-factor 1
```

Messages published to `ssi_ps` must conform to
`proto/markets/v1/ssi_ps.proto` (`markets.v1.SsiPsSnapshot`). The producer CLI
reads line-delimited payloads from `apps/producer/internal/data` (override with
`PRODUCER_DATA_FILE`), parses them via the Hose stock parser, and writes
protobuf-encoded snapshots. Key environment variables:

- `PRODUCER_KAFKA_BROKERS` – comma separated broker list (default `localhost:9092`).
- `PRODUCER_TOPIC` – Kafka topic name (default `ssi_ps`).
- `PRODUCER_TOPIC_PARTITIONS` / `PRODUCER_TOPIC_REPLICATION` – optional topic sizing.
- `PRODUCER_DATA_FILE` – alternate input file path.

`make consumer` consumes the same topic, decodes the protobuf payload, and
stores each full snapshot as JSON inside a Redis sorted set keyed by
`ssi_ps:<code>`, while also publishing to a Redis Stream
(`ssi_ps_stream:<code>`) for live subscribers. Query a time window via
`ZRANGEBYSCORE ssi_ps:<code> <start_ms> <end_ms>` and load the JSON rows into a
dataframe. Configure with:

- `CONSUMER_KAFKA_BROKERS` – broker list.
- `CONSUMER_KAFKA_TOPIC` / `CONSUMER_KAFKA_GROUP` – topic & consumer group.
- `CONSUMER_REDIS_ADDR` – Redis endpoint.
- `CONSUMER_REDIS_KEY_FMT` – optional `fmt.Sprintf` pattern for the sorted-set key.
- `CONSUMER_REDIS_STREAM_FMT` – optional `fmt.Sprintf` pattern for stream keys.
- `CONSUMER_METRIC_PREFIX` – optional namespace applied when no custom format is supplied.

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
