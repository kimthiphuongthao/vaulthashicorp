-- Create legacy schema/table used by the SQL updater demo

CREATE TABLE IF NOT EXISTS users (
  username VARCHAR(255) PRIMARY KEY,
  password_hash TEXT NOT NULL
);

INSERT IGNORE INTO users (username, password_hash) VALUES
  ('alice', ''),
  ('bob', ''),
  ('charlie', '');
