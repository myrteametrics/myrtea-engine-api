-- +goose Up
-- +goose StatementBegin
CREATE TABLE api_keys
(
    id           uuid PRIMARY KEY,
    key_hash     text         NOT NULL,
    key_prefix   varchar(8)   NOT NULL,
    name         varchar(100) NOT NULL,
    role_id      uuid         NOT NULL,
    created_at   timestamp    NOT NULL DEFAULT NOW(),
    expires_at   timestamp    NULL,
    last_used_at timestamp    NULL,
    is_active    boolean      NOT NULL DEFAULT TRUE,
    created_by   varchar(100) NOT NULL,
    CONSTRAINT fk_role FOREIGN KEY (role_id) REFERENCES roles_v4 (id)
);

-- Composite index for frequent searches
CREATE INDEX idx_api_keys_search ON api_keys (key_prefix, is_active, expires_at);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_api_keys_search;

DROP TABLE IF EXISTS api_keys;
-- +goose StatementEnd