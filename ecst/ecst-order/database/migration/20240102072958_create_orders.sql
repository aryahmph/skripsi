-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders
(
    id         VARCHAR(26) NOT NULL PRIMARY KEY,
    ticket_id  VARCHAR(26) NOT NULL REFERENCES tickets (id),
    user_id    VARCHAR(26) NOT NULL,
    payment_id VARCHAR(26),
    status     VARCHAR(15) NOT NULL,
    amount     BIGINT      NOT NULL,
    version    INT         NOT NULL DEFAULT 1,
    created_at TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX orders_unique ON orders (ticket_id, status) WHERE NOT (status = 'expired');
CREATE INDEX orders_user_id_idx ON orders (user_id);
CREATE INDEX orders_payment_id_idx ON orders (payment_id);
CREATE INDEX orders_status_idx ON orders (status);
CREATE INDEX orders_version_idx ON orders (version);
CREATE INDEX orders_deleted_at_idx ON orders (deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
