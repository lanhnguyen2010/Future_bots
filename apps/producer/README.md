# Producer Service

Lightweight HTTP facade for publishing Kafka messages. The service exposes a
`POST /api/v1/messages` endpoint that accepts payloads and pushes them to a
configured topic on the local Kafka cluster (provisioned via `docker-compose`).

## Configuration

Environment variable | Description | Default
-------------------- | ----------- | -------
`PRODUCER_ADDR` | HTTP listen address | `:8090`
`PRODUCER_SHUTDOWN_TIMEOUT` | Graceful shutdown timeout | `10s`
`PRODUCER_KAFKA_BROKERS` | Comma-separated broker list | `localhost:9092`
`PRODUCER_DEFAULT_TOPIC` | Fallback topic when the request omits one | _(required if request omits topic)_

## API

```http
POST /api/v1/messages
Content-Type: application/json
{
  "topic": "orders",          // optional when PRODUCER_DEFAULT_TOPIC set
  "key": "client-1",          // optional string key
  "value": "{...}",           // required payload (string)
  "headers": {                 // optional string headers
    "source": "demo"
  }
}
```

A successful request returns `202 Accepted` with `{"status":"queued"}`. The
service responds with `400` when validation fails (e.g. missing `value` or
missing topic) and `500` for Kafka write errors.

## Running Locally

```bash
PRODUCER_DEFAULT_TOPIC=bot-orders \
PRODUCER_KAFKA_BROKERS=localhost:9094 \
go run ./cmd/producer
```

Use `curl` or similar to push messages:

```bash
curl -X POST http://localhost:8090/api/v1/messages \
  -H 'Content-Type: application/json' \
  -d '{"value":"{\"event\":\"ping\"}","headers":{"source":"curl"}}'
```

The service shares the logging/shutdown utilities from `libs/go/platform` and
can be containerised alongside the rest of the stack when needed.
