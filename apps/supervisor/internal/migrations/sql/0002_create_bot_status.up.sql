CREATE TABLE IF NOT EXISTS bot_status (
    bot_id UUID PRIMARY KEY,
    phase TEXT NOT NULL,
    reason TEXT,
    image_running TEXT,
    last_heartbeat TIMESTAMPTZ,
    p95_tick_ms DOUBLE PRECISION,
    intents_per_s DOUBLE PRECISION,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
