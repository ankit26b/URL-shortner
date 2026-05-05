CREATE TABLE IF NOT EXISTS url_clicks (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(64) NOT NULL,
    clicked_at TIMESTAMPTZ NOT NULL,
    ip_address VARCHAR(64) NOT NULL,
    user_agent TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_url_clicks_short_code ON url_clicks (short_code);
CREATE INDEX IF NOT EXISTS idx_url_clicks_clicked_at ON url_clicks (clicked_at);
