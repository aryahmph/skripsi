-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS payments
(
    id         VARCHAR(26) NOT NULL PRIMARY KEY,
    order_id   VARCHAR(26) NOT NULL UNIQUE REFERENCES orders (id),
    created_at TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX payments_deleted_at_idx ON payments (deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS payments;
-- +goose StatementEnd
