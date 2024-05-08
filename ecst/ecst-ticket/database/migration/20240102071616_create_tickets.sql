-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tickets
(
    id              VARCHAR(26) NOT NULL PRIMARY KEY,
    ticket_group_id VARCHAR(26) NOT NULL REFERENCES ticket_groups (id),
    order_id        VARCHAR(26) UNIQUE,
    code            VARCHAR(50) NOT NULL,
    category        VARCHAR(50) NOT NULL,
    price           BIGINT      NOT NULL,
    version         INT         NOT NULL DEFAULT 1,
    created_at      TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP,
    UNIQUE (ticket_group_id, code)
);

CREATE INDEX tickets_order_id_idx ON tickets (order_id);
CREATE INDEX tickets_category_idx ON tickets (category);
CREATE INDEX tickets_price_idx ON tickets (price);
CREATE INDEX tickets_version_idx ON tickets (version);
CREATE INDEX tickets_deleted_at_idx ON tickets (deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tickets;
-- +goose StatementEnd
