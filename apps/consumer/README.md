# Consumer Service

Kafka consumer that reads `markets.v1.SsiPsSnapshot` messages from the `ssi_ps`
topic and stores complete snapshots in Redis:

- `ZRANGEBYSCORE`-ready sorted sets keyed by code (`ssi_ps:<code>`) containing the
  JSON payload for each tick.
- Redis Streams (`ssi_ps_stream:<code>`) so downstream bots can subscribe to
  live snapshots.

## Configuration

Environment variable | Description | Default
-------------------- | ----------- | -------
`CONSUMER_KAFKA_BROKERS` | Comma-separated broker list | `localhost:9092`
`CONSUMER_KAFKA_TOPIC`   | Kafka topic to consume     | `ssi_ps`
`CONSUMER_KAFKA_GROUP`   | Consumer group id          | `ssi_ps_consumer`
`CONSUMER_REDIS_ADDR`    | Redis endpoint             | `localhost:6379`
`CONSUMER_REDIS_KEY_FMT`    | Sorted-set key pattern (`fmt.Sprintf`) | `ssi_ps:%s`
`CONSUMER_REDIS_STREAM_FMT` | Stream key pattern (`fmt.Sprintf`)     | `ssi_ps_stream:%s`
`CONSUMER_METRIC_PREFIX`    | Namespace used when key fmt omitted    | `ssi_ps`

Sorted-set members are JSON strings, so range queries return rows suitable for
dataframe ingestion. Streams carry the same payload for real-time consumers.

## Running Locally

```bash
CONSUMER_KAFKA_BROKERS=localhost:9092 \
CONSUMER_REDIS_ADDR=localhost:6379 \
make consumer
```

Ensure the producer (`make producer`) is publishing snapshots so the consumer
has data to ingest.
