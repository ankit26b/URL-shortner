CREATE INDEX IF NOT EXISTS idx_urls_short_code_active
ON urls (short_code)
WHERE expires_at IS NULL OR expires_at > NOW();

CREATE INDEX IF NOT EXISTS idx_urls_expires_at
ON urls (expires_at);

CREATE INDEX IF NOT EXISTS idx_urls_created_at
ON urls (created_at DESC);
