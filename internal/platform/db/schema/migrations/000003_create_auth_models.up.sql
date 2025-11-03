CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    jti UUID UNIQUE NOT NULL,
    session_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);

CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens (token_hash);

CREATE INDEX idx_refresh_tokens_jti ON refresh_tokens (jti);

CREATE INDEX idx_refresh_tokens_revoked ON refresh_tokens (revoked);

CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens (expires_at);