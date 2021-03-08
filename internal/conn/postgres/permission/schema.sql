--
-- Automated script, we do not need NOTICE and WARNING
--
SET client_min_messages TO ERROR;

CREATE TABLE IF NOT EXISTS permission_principals (
  -- IDs are not really human language text, so set them to be
  -- COLLATE "C" rather than the DB default collation.
  user_id TEXT COLLATE "C" NOT NULL,
  principal TEXT NOT NULL,

  PRIMARY KEY (user_id, principal)
);

CREATE TABLE IF NOT EXISTS permission_paths (
  path TEXT COLLATE "C" NOT NULL,
  permission TEXT NOT NULL,
  principal TEXT NOT NULL,

  PRIMARY KEY (path, permission, principal)
);
CREATE INDEX IF NOT EXISTS idx_permission_paths_permission
  ON permission_paths(permission);
CREATE INDEX IF NOT EXISTS idx_permission_paths_principal
  ON permission_paths(principal);

-- Same table as exists in the storage backend, but used to track
-- migration status for both. Only one schema actually has to create
-- it.
CREATE TABLE IF NOT EXISTS metainfo (
  name VARCHAR(128) NOT NULL,
  value VARCHAR(512) NOT NULL,

  PRIMARY KEY (name)
);
INSERT INTO metainfo VALUES ('permission_schema_version', '1')
ON CONFLICT (name) DO NOTHING;
