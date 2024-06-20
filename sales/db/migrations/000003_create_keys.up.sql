CREATE TABLE IF NOT EXISTS keys (
    kid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expiry TIMESTAMP NOT NULL,
    signing_key TEXT NOT NULL,
    public_key TEXT NOT NULL,
    encryption_key TEXT NOT NULL,
    revoked BOOLEAN DEFAULT FALSE
);
