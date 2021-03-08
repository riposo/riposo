--
-- Automated script, we do not need NOTICE and WARNING
--
SET client_min_messages TO ERROR;

--
-- Actually stored objects.
--
CREATE TABLE IF NOT EXISTS storage_objects (
    -- Node paths and object IDs are stored as text, and not human language.
    -- Therefore, we store them in the C collation. This lets Postgres
    -- use the index for prefix matching (path LIKE
    -- '/buckets/abc/%').
    path TEXT COLLATE "C" NOT NULL,
    id TEXT COLLATE "C" NOT NULL,

    -- Timestamp as millisecond epoch.
    last_modified BIGINT NOT NULL,

    -- JSONB, 2x faster than JSON.
    data JSONB NOT NULL DEFAULT '{}'::JSONB,

    deleted BOOLEAN NOT NULL,

    PRIMARY KEY (path, id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_storage_objects_path_last_modified
    ON storage_objects(path, last_modified DESC);
CREATE INDEX IF NOT EXISTS idx_storage_objects_last_modified
    ON storage_objects(last_modified);

--
-- Create node timestamps
--
CREATE TABLE IF NOT EXISTS storage_timestamps (
  -- The node path.
  path TEXT COLLATE "C" NOT NULL,

  -- Timestamp as millisecond epoch.
  last_modified BIGINT NOT NULL,

  PRIMARY KEY (path)
);

--
-- Convert timestamps to milliseconds epoch integer
--
CREATE OR REPLACE FUNCTION as_epoch(ts TIMESTAMP) RETURNS BIGINT AS $$
BEGIN
  RETURN (EXTRACT(EPOCH FROM ts) * 1000)::BIGINT;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

--
-- Set or increment the current object node epoch.
--
CREATE OR REPLACE FUNCTION storage_timestamps_increment(node_path TEXT)
RETURNS BIGINT AS $$
DECLARE
  epoch BIGINT;
BEGIN
  epoch := NULL;
  WITH incremented AS (
    INSERT INTO storage_timestamps (path, last_modified)
    VALUES (node_path, as_epoch(clock_timestamp()::TIMESTAMP))
    ON CONFLICT (path) DO UPDATE
    SET
      last_modified = CASE
        WHEN storage_timestamps.last_modified < EXCLUDED.last_modified
        THEN EXCLUDED.last_modified
        ELSE storage_timestamps.last_modified + 1
        END
    RETURNING last_modified
  ) SELECT last_modified INTO epoch FROM incremented;
  RETURN epoch;
END;
$$ LANGUAGE plpgsql VOLATILE;

--
-- Trigger function
--
-- This increments the storage_timestamps and uses the resulting
-- timestamp as the last_modified on the new object to avoid
-- clashes.
--
CREATE OR REPLACE FUNCTION storage_objects_set_last_modified()
RETURNS trigger AS $$
BEGIN
  SELECT storage_timestamps_increment(NEW.path) INTO NEW.last_modified;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

--
-- Triggers to set last_modified on INSERT/UPDATE
--

DROP TRIGGER IF EXISTS tgr_storage_objects_last_modified ON storage_objects;

CREATE TRIGGER tgr_storage_objects_last_modified
BEFORE INSERT OR UPDATE OF data, deleted ON storage_objects
FOR EACH ROW EXECUTE PROCEDURE storage_objects_set_last_modified();

--
-- metainfo table
--
CREATE TABLE IF NOT EXISTS metainfo (
  name VARCHAR(128) NOT NULL,
  value VARCHAR(512) NOT NULL,

  PRIMARY KEY (name)
);
INSERT INTO metainfo VALUES ('storage_schema_version', '1')
ON CONFLICT (name) DO NOTHING;
