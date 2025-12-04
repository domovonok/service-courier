-- +goose Up
-- +goose StatementBegin
ALTER TABLE couriers
    ADD COLUMN transport_type TEXT NOT NULL DEFAULT 'on_foot';

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
ALTER TABLE couriers
    DROP COLUMN transport_type;

DROP TABLE IF EXISTS delivery;
-- +goose StatementEnd

