-- TimescaleDB schema excerpt for Qubit Bot Trading Platform

CREATE TABLE IF NOT EXISTS desired_bots(
  bot_id UUID PRIMARY KEY,
  account_id TEXT NOT NULL,
  name TEXT NOT NULL,
  image TEXT NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT FALSE,
  config JSONB NOT NULL,
  config_rev INTEGER NOT NULL DEFAULT 1,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS bot_status(
  bot_id UUID PRIMARY KEY,
  phase TEXT NOT NULL,
  reason TEXT,
  image_running TEXT,
  last_heartbeat TIMESTAMPTZ,
  p95_tick_ms DOUBLE PRECISION,
  intents_per_s DOUBLE PRECISION,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ticks(
  ts timestamptz NOT NULL,
  symbol text NOT NULL,
  price numeric NOT NULL,
  volume numeric NOT NULL,
  PRIMARY KEY(ts, symbol)
);

-- SELECT create_hypertable('ticks', by_range('ts'), if_not_exists => TRUE);

CREATE TABLE IF NOT EXISTS orders(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  bot_id uuid NOT NULL,
  symbol text NOT NULL,
  side text NOT NULL CHECK (side in ('buy','sell')),
  qty numeric NOT NULL,
  price numeric,
  status text NOT NULL DEFAULT 'new',
  provider_order_id text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS executions(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id uuid NOT NULL REFERENCES orders(id),
  fill_qty numeric NOT NULL,
  fill_price numeric NOT NULL,
  fee numeric NOT NULL DEFAULT 0,
  filled_at timestamptz NOT NULL DEFAULT now()
);
