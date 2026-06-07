CREATE TABLE IF NOT EXISTS orders (
    id         BIGSERIAL    PRIMARY KEY,
    order_id   VARCHAR(36)  NOT NULL UNIQUE,
    user_name  VARCHAR(100) NOT NULL,
    phone_num  VARCHAR(20)  NOT NULL,
    area       INT          NOT NULL,
    status     VARCHAR(20)  NOT NULL DEFAULT 'success',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (user_name, area)
);

CREATE INDEX IF NOT EXISTS idx_orders_user ON orders (user_name);
CREATE INDEX IF NOT EXISTS idx_orders_area ON orders (area);

ALTER TABLE orders
ADD CONSTRAINT orders_user_name_area_key
UNIQUE (user_name, area);