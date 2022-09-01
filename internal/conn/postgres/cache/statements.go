package cache

// Placeholders:
//
//	$1 - key
//	$2 - now
const sqlGetKey = `
SELECT value
FROM cache_keys
WHERE key = $1
  AND expires_at > $2
`

// Placeholders:
//
//	$1 - key
//	$2 - val
//	$3 - exp
const sqlSetKey = `
INSERT INTO cache_keys (key, value, expires_at)
VALUES ($1, $2, $3)
ON CONFLICT (key) DO UPDATE SET
  value = EXCLUDED.value,
  expires_at = EXCLUDED.expires_at
`

// Placeholders:
//
//	$1 - key
//	$2 - now
const sqlDelKey = `
DELETE FROM cache_keys
WHERE key = $1
  AND expires_at > $2
RETURNING key
`

// Placeholders:
//
//	$1 - now
const sqlPrune = `
DELETE FROM cache_keys
WHERE expires_at <= $1
`
