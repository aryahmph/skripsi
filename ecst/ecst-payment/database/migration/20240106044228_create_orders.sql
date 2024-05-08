-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders
(
    id         VARCHAR(26) NOT NULL PRIMARY KEY,
    amount     BIGINT      NOT NULL,
    status     VARCHAR(15) NOT NULL,
    version    INT         NOT NULL DEFAULT 1,
    created_at TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX orders_status_idx ON orders (status);
CREATE INDEX orders_version_idx ON orders (version);
CREATE INDEX orders_deleted_at_idx ON orders (deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
