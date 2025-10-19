# Protobuf schemas

This directory hosts the canonical Protocol Buffer schemas for the Kafka topics
that connect trading bots with the control plane, executor, and risk services.

## Topics

| Topic Pattern | Schema | Description |
| ------------- | ------ | ----------- |
| `orders.intent.account.<account_id>.<bot_id>` | [`orders/v1/orders.proto`](orders/v1/orders.proto) (`OrderIntent`) | Trading bot order intents produced to Kafka. |
| `orders.event.account.<account_id>.<bot_id>` | [`orders/v1/orders.proto`](orders/v1/orders.proto) (`OrderEvent`) | Execution acknowledgements, fills, rejections, and cancels emitted by the executor. |
| `risk.alerts.account.<account_id>` | [`risk/v1/alerts.proto`](risk/v1/alerts.proto) (`RiskAlert`) | Broadcast risk policy alerts for supervisory dashboards and bots. |
| `bot.commands.<bot_id>` | [`bot/v1/commands.proto`](bot/v1/commands.proto) (`BotCommandEnvelope`) | Supervisor-issued runtime commands (start, stop, rollout). |
| `ssi_ps` | [`markets/v1/ssi_ps.proto`](markets/v1/ssi_ps.proto) (`SsiPsSnapshot`) | Hose PowerScreen market depth snapshots parsed from SSI feed. |

## Generating code

The repository uses `protoc` with language-specific plugins. Inside the dev
container (see [`.devcontainer`](../.devcontainer)), install the desired plugins
and run, for example:

```bash
mkdir -p gen/go
protoc \
  --go_out=gen/go --go_opt=paths=source_relative \
  --go-grpc_out=gen/go --go-grpc_opt=paths=source_relative \
  $(find proto -name '*.proto')
```

Add additional plugin options (TypeScript, Python, etc.) as needed for other
components of the monorepo.
