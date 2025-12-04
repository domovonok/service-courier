-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS couriers (
    id             BIGSERIAL PRIMARY KEY,
    name           TEXT NOT NULL,
    phone          TEXT NOT NULL UNIQUE,
    status         TEXT NOT NULL DEFAULT 'available',
    created_at     TIMESTAMP DEFAULT NOW(),
    updated_at     TIMESTAMP DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS couriers;
-- +goose StatementEnd
