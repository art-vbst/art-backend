DROP INDEX IF EXISTS idx_refresh_tokens_revoked;

DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;

DROP INDEX IF EXISTS idx_refresh_tokens_token_hash;

DROP INDEX IF EXISTS idx_refresh_tokens_user_id;

DROP TABLE IF EXISTS refresh_tokens;

DROP TABLE IF EXISTS users;