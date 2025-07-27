BEGIN;

CREATE TABLE IF NOT EXISTS users (    
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('viewer', 'editor')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users (username);

INSERT INTO users (username, password, role) VALUES
('ted', '$2a$12$4VS2sXb5iPx/Thy2YBCDfe09CT5zFKBb68A8xiC8TtjYRE1jGaXLG', 'viewer'),
('mike', '$2a$12$vhaVXeHEWYix8mewZio3PechUVn8NSQYkBVEsWTs0vAHeXtJXvRzq', 'viewer'),
('john', '$2a$12$I69KxNA/5IZF0lutR4wGl.i2ZxjpAp1QhAujnIZ0Z9roEyhLOs8BO', 'editor'),
('uncle_bob', '$2a$12$FHTITVO0Gz4xoTqsvUXIP.srslqQNqdQK5NaOuzg0yS9TkRfaA9G2', 'editor')
ON CONFLICT (username) DO NOTHING;

CREATE TABLE IF NOT EXISTS feature_flags (
    id UUID PRIMARY KEY NOT NULL,
    key TEXT NOT NULL,
    description TEXT NOT NULL,
    enabled BOOLEAN NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMIT;