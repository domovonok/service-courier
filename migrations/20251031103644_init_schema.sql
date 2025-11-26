-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS couriers (
    id             BIGSERIAL PRIMARY KEY,
    name           TEXT NOT NULL,
    phone          TEXT NOT NULL UNIQUE,
    status         TEXT NOT NULL DEFAULT 'available',
    transport_type TEXT NOT NULL DEFAULT 'on_foot',
    created_at     TIMESTAMP DEFAULT NOW(),
    updated_at     TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS delivery (
    id          BIGSERIAL PRIMARY KEY,
    courier_id  BIGINT NOT NULL,
    order_id    VARCHAR(255) NOT NULL,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deadline    TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS couriers;
DROP TABLE IF EXISTS delivery;
-- +goose StatementEnd
