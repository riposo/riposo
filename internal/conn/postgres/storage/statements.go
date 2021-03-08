package storage

// Placeholders:
//   $1 - path
const sqlGetModTime = `
SELECT last_modified
FROM storage_timestamps
WHERE path = $1
`

// Placeholders:
//   $1 - path
//   $2 - id
const sqlExistsObject = `
SELECT TRUE
FROM storage_objects
WHERE path = $1
  AND id = $2
  AND NOT deleted
LIMIT 1
`

// Placeholders:
//   $1 - path
//   $2 - id
const sqlGetObject = `
SELECT last_modified, data
FROM storage_objects
WHERE path = $1
  AND id = $2
  AND NOT deleted
`

// Placeholders:
//   $1 - path
//   $2 - id
const sqlGetObjectForUpdate = `
SELECT last_modified, data
FROM storage_objects
WHERE path = $1
  AND NOT deleted
  AND id = $2
FOR UPDATE
`

// Placeholders:
//   $1 - path
//   $2 - id
//   $3 - data
const sqlCreateObject = `
INSERT INTO storage_objects (
  path,
  id,
  data,
  last_modified,
  deleted
)
SELECT
  $1,
  $2,
  ($3)::JSONB,
  NULL,
  FALSE
ON CONFLICT (path, id) DO UPDATE SET
  data = EXCLUDED.data,
  last_modified = EXCLUDED.last_modified,
  deleted = FALSE
WHERE storage_objects.deleted = TRUE
RETURNING last_modified
`

// Placeholders:
//   $1 - path
//   $2 - id
//   $3 - data
const sqlUpdateObject = `
INSERT INTO storage_objects (
  path,
  id,
  data,
  last_modified,
  deleted
)
VALUES (
  $1,
  $2,
  ($3)::JSONB,
  NULL,
  FALSE
)
ON CONFLICT (path, id) DO UPDATE SET
  data = EXCLUDED.data,
  last_modified = EXCLUDED.last_modified,
  deleted = FALSE
RETURNING last_modified
`

// Placeholders:
//   $1 - path
//   $2 - id
const sqlDeleteObject = `
UPDATE storage_objects
  SET deleted = TRUE
WHERE path = $1
AND id = $2
AND NOT deleted
RETURNING last_modified, data
`

// Placeholders:
//   $1 - path pattern
const sqlDeleteObjectNested = `
UPDATE storage_objects
  SET deleted = TRUE
WHERE path LIKE $1
AND NOT deleted
`

// Placeholders:
//   $1 - deleteAll?
//   $2 - olderThan
const sqlPurgeObjects = `
DELETE FROM storage_objects
WHERE deleted
  AND ($1 OR last_modified < $2)
`
