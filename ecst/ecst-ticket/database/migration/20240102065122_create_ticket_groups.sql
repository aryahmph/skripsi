-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS ticket_groups
(
    id         VARCHAR(26)  NOT NULL PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX ticket_groups_name_idx ON ticket_groups (name);
CREATE INDEX ticket_groups_deleted_at_idx ON ticket_groups (deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ticket_groups;
-- +goose StatementEnd
