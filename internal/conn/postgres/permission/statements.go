package permission

// Placeholders:
//   $1 - userID
//   $2 - 'system.Authenticated'
//   $3 - 'system.Everyone'
const sqlGetUserPrincipals = `
SELECT user_id, principal
FROM permission_principals
WHERE user_id IN ($1, $2, $3)
`

// Placeholders:
//   $1 - principal
//   $2 - userIDs
const sqlRemoveUserPrincipal = `
DELETE FROM permission_principals
WHERE principal = $1
  AND user_id = ANY($2)
`

// Placeholders:
//   $1 - principals
const sqlPurgeUserPrincipals = `
DELETE FROM permission_principals
WHERE principal = ANY($1)
`

// Placeholders:
//   $1 - path
//   $2 - perm
const sqlGetACEPrincipals = `
SELECT principal
FROM permission_paths
WHERE path = $1
  AND permission = $2
`

// Placeholders:
//   $1 - path
//   $2 - perm
const sqlMatchACEPrincipals = `
SELECT principal
FROM permission_paths
WHERE path ~ $1
  AND permission = $2
`

// Placeholders:
//   $1 - path
//   $2 - perm
//   $3 - principal
const sqlInsertACE = `
INSERT INTO permission_paths (path, permission, principal)
VALUES ($1, $2, $3)
ON CONFLICT (path, permission, principal) DO NOTHING
`

// Placeholders:
//   $1 - path
//   $2 - perm
//   $3 - principal
const sqlDeleteACE = `
DELETE FROM permission_paths
WHERE path = $1
  AND permission = $2
  AND principal = $3
`

// Placeholders:
//   $1 - path
const sqlGetPerms = `
SELECT permission, principal
FROM permission_paths
WHERE path = $1
`
