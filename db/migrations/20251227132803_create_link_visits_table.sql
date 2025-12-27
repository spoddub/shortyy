-- +goose Up
CREATE TABLE IF NOT EXISTS link_visits (
                                           id         BIGSERIAL PRIMARY KEY,
                                           link_id    BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    ip         TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    referer    TEXT NOT NULL DEFAULT '',
    status     INT  NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_link_visits_link_id ON link_visits(link_id);
CREATE INDEX IF NOT EXISTS idx_link_visits_created_at ON link_visits(created_at);

-- +goose Down
DROP TABLE IF EXISTS link_visits;
