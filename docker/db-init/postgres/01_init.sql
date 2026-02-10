-- Create legacy schema/table used by the SQL updater demo

CREATE TABLE IF NOT EXISTS users (
  username TEXT PRIMARY KEY,
  password_hash TEXT NOT NULL DEFAULT ''
);

INSERT INTO users (username) VALUES ('alice') ON CONFLICT DO NOTHING;
INSERT INTO users (username) VALUES ('bob') ON CONFLICT DO NOTHING;
INSERT INTO users (username) VALUES ('charlie') ON CONFLICT DO NOTHING;
