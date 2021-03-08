--
-- Automated script, we do not need NOTICE and WARNING
--
SET client_min_messages TO ERROR;

CREATE TABLE IF NOT EXISTS cache_keys (
  -- Keys are not really human language text, so set them to be
  -- COLLATE "C" rather than the DB default collation.
  key VARCHAR(256) PRIMARY KEY,
  value TEXT NOT NULL,
  expires_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_cache_keys_expires_at ON cache_keys(expires_at);

-- Same table as exists in the storage backend, but used to track
-- migration status for both. Only one schema actually has to create
-- it.
CREATE TABLE IF NOT EXISTS metainfo (
  name VARCHAR(128) NOT NULL,
  value VARCHAR(512) NOT NULL,

  PRIMARY KEY (name)
);
INSERT INTO metainfo VALUES ('cache_schema_version', '1')
ON CONFLICT (name) DO NOTHING;
