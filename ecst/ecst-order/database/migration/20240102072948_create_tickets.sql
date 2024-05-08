-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tickets
(
    id         VARCHAR(26) NOT NULL PRIMARY KEY,
    code       VARCHAR(50) NOT NULL,
    category   VARCHAR(50) NOT NULL,
    price      BIGINT      NOT NULL,
    version    INT         NOT NULL,
    created_at TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX tickets_version_idx ON tickets (version);
CREATE INDEX tickets_deleted_at_idx ON tickets (deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
