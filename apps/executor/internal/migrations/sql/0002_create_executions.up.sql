CREATE TABLE IF NOT EXISTS executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id),
    fill_qty NUMERIC NOT NULL,
    fill_price NUMERIC NOT NULL,
    fee NUMERIC NOT NULL DEFAULT 0,
    filled_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
