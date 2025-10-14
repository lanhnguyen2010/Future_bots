CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bot_id UUID NOT NULL,
    symbol TEXT NOT NULL,
    side TEXT NOT NULL CHECK (side in ('buy','sell')),
    qty NUMERIC NOT NULL,
    price NUMERIC,
    status TEXT NOT NULL DEFAULT 'new',
    provider_order_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
