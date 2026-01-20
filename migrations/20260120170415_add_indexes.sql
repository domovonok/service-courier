-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_delivery_courier_id ON delivery(courier_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_delivery_order_id ON delivery(order_id);
CREATE INDEX IF NOT EXISTS idx_delivery_deadline ON delivery(deadline);
CREATE INDEX IF NOT EXISTS idx_couriers_status ON couriers(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_delivery_courier_id;
DROP INDEX IF EXISTS idx_delivery_order_id;
DROP INDEX IF EXISTS idx_delivery_deadline;
DROP INDEX IF EXISTS idx_couriers_status;
-- +goose StatementEnd
