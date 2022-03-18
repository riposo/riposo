package permission

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"net/url"
	"strings"

	"github.com/bsm/minisql"
	"github.com/lib/pq"
	"github.com/riposo/riposo/internal/conn/postgres/common"
	"github.com/riposo/riposo/pkg/conn/permission"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"
	"github.com/riposo/riposo/pkg/util"
	"go.uber.org/multierr"
)

const schemaVersion = 1

//go:embed schema.sql
var embedFS embed.FS

func init() {
	permission.Register("postgres", func(ctx context.Context, uri *url.URL, _ riposo.Helpers) (permission.Backend, error) {
		return Connect(ctx, uri.String())
	})
}

type conn struct {
	db   *sql.DB
	stmt struct {
		getUserPrincipals, removeUserPrincipal, purgeUserPrincipals,
		getACEPrincipals, matchACEPrincipals,
		insertACE, deleteACE,
		getPerms, deletePerms *sql.Stmt
	}
}

// Connect connects to a PostgreSQL server.
func Connect(ctx context.Context, dsn string) (permission.Backend, error) {
	// Connect to the DB.
	db, err := common.Connect(ctx, dsn, "permission_schema_version", schemaVersion, embedFS)
	if err != nil {
		return nil, err
	}

	// Create connection struct, prepare statements.
	cn := &conn{db: db}
	if err := cn.prepare(ctx); err != nil {
		_ = cn.Close()
	}
	return cn, nil
}

//nolint:sqlclosecheck
func (cn *conn) prepare(ctx context.Context) (err error) {
	if cn.stmt.getUserPrincipals, err = cn.db.PrepareContext(ctx, sqlGetUserPrincipals); err != nil {
		return
	}
	if cn.stmt.getACEPrincipals, err = cn.db.PrepareContext(ctx, sqlGetACEPrincipals); err != nil {
		return
	}
	if cn.stmt.matchACEPrincipals, err = cn.db.PrepareContext(ctx, sqlMatchACEPrincipals); err != nil {
		return
	}
	if cn.stmt.removeUserPrincipal, err = cn.db.PrepareContext(ctx, sqlRemoveUserPrincipal); err != nil {
		return
	}
	if cn.stmt.purgeUserPrincipals, err = cn.db.PrepareContext(ctx, sqlPurgeUserPrincipals); err != nil {
		return
	}
	if cn.stmt.insertACE, err = cn.db.PrepareContext(ctx, sqlInsertACE); err != nil {
		return
	}
	if cn.stmt.deleteACE, err = cn.db.PrepareContext(ctx, sqlDeleteACE); err != nil {
		return
	}
	if cn.stmt.getPerms, err = cn.db.PrepareContext(ctx, sqlGetPerms); err != nil {
		return
	}
	return
}

// Ping implements permission.Backend interface.
func (cn *conn) Ping(ctx context.Context) error {
	return cn.db.PingContext(ctx)
}

// Begin implements permission.Backend interface.
func (cn *conn) Begin(ctx context.Context) (permission.Transaction, error) {
	tx, err := cn.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &transaction{Tx: tx, cn: cn, ctx: ctx}, nil
}

// Close implements permission.Backend.
func (cn *conn) Close() (err error) {
	if cn.stmt.getUserPrincipals != nil {
		err = multierr.Append(err, cn.stmt.getUserPrincipals.Close())
	}
	if cn.stmt.removeUserPrincipal != nil {
		err = multierr.Append(err, cn.stmt.removeUserPrincipal.Close())
	}
	if cn.stmt.purgeUserPrincipals != nil {
		err = multierr.Append(err, cn.stmt.purgeUserPrincipals.Close())
	}
	if cn.stmt.getACEPrincipals != nil {
		err = multierr.Append(err, cn.stmt.getACEPrincipals.Close())
	}
	if cn.stmt.matchACEPrincipals != nil {
		err = multierr.Append(err, cn.stmt.matchACEPrincipals.Close())
	}
	if cn.stmt.insertACE != nil {
		err = multierr.Append(err, cn.stmt.insertACE.Close())
	}
	if cn.stmt.deleteACE != nil {
		err = multierr.Append(err, cn.stmt.deleteACE.Close())
	}
	if cn.stmt.getPerms != nil {
		err = multierr.Append(err, cn.stmt.getPerms.Close())
	}
	if cn.db != nil {
		err = multierr.Append(err, cn.db.Close())
	}
	return
}

// --------------------------------------------------------------------

type transaction struct {
	*sql.Tx
	cn  *conn
	ctx context.Context
}

// Commit implements permission.Transaction interface.
func (tx *transaction) Commit() error {
	return normErr(tx.Tx.Commit())
}

// Rollback implements permission.Transaction interface.
func (tx *transaction) Rollback() error {
	return normErr(tx.Tx.Rollback())
}

// Flush implements permission.Transaction interface.
func (tx *transaction) Flush() error {
	_, err := tx.ExecContext(tx.ctx, `TRUNCATE permission_paths, permission_principals`)
	return normErr(err)
}

// GetUserPrincipals implements permission.Transaction.
func (tx *transaction) GetUserPrincipals(userID string) ([]string, error) {
	stmt := tx.StmtContext(tx.ctx, tx.cn.stmt.getUserPrincipals)
	defer stmt.Close()

	rows, err := stmt.QueryContext(tx.ctx, userID, riposo.Authenticated, riposo.Everyone)
	if err != nil {
		return nil, normErr(err)
	}
	defer rows.Close()

	res := util.NewSet()
	switch userID {
	default:
		res.Add(riposo.Authenticated)
		fallthrough
	case riposo.Authenticated:
		res.Add(riposo.Everyone)
		fallthrough
	case riposo.Everyone:
		res.Add(userID)
	}

	for rows.Next() {
		var rowUserID, rowPrincipal string
		if err := rows.Scan(&rowUserID, &rowPrincipal); err != nil {
			return nil, err
		}

		switch userID {
		case riposo.Everyone:
			if rowUserID == riposo.Everyone {
				res.Add(rowPrincipal)
			}
		default:
			res.Add(rowPrincipal)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res.Slice(), nil
}

// AddUserPrincipal implements permission.Transaction.
func (tx *transaction) AddUserPrincipal(principal string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	stmt := minisql.Pooled()
	defer minisql.Release(stmt)

	stmt.AppendString("INSERT INTO permission_principals (user_id, principal) VALUES ")
	for i, userID := range userIDs {
		if i > 0 {
			stmt.AppendByte(',')
		}
		stmt.AppendByte('(')
		stmt.AppendValue(userID)
		stmt.AppendByte(',')
		stmt.AppendValue(principal)
		stmt.AppendByte(')')
	}
	stmt.AppendString(" ON CONFLICT (user_id, principal) DO NOTHING")

	_, err := stmt.ExecContext(tx.ctx, tx)
	return normErr(err)
}

// RemoveUserPrincipal implements permission.Transaction.
func (tx *transaction) RemoveUserPrincipal(principal string, userIDs []string) (err error) {
	if len(userIDs) == 0 {
		return
	}

	stmt := tx.StmtContext(tx.ctx, tx.cn.stmt.removeUserPrincipal)
	defer stmt.Close()

	_, err = stmt.
		ExecContext(tx.ctx, principal, pq.Array(userIDs))
	return normErr(err)
}

// PurgeUserPrincipals implements permission.Transaction.
func (tx *transaction) PurgeUserPrincipals(principals []string) error {
	if len(principals) == 0 {
		return nil
	}

	stmt := tx.StmtContext(tx.ctx, tx.cn.stmt.purgeUserPrincipals)
	defer stmt.Close()

	_, err := stmt.ExecContext(tx.ctx, pq.Array(principals))
	return normErr(err)
}

// GetACEPrincipals implements permission.Transaction.
func (tx *transaction) GetACEPrincipals(ent permission.ACE) ([]string, error) {
	stmt := tx.cn.stmt.getACEPrincipals
	if ent.Path.IsNode() {
		stmt = tx.cn.stmt.matchACEPrincipals
	}
	stmt = tx.StmtContext(tx.ctx, stmt)
	defer stmt.Close()

	return scanStringSlice(stmt.QueryContext(tx.ctx, ent.Path, ent.Perm))
}

// AddACEPrincipal implements permission.Transaction.
func (tx *transaction) AddACEPrincipal(principal string, ent permission.ACE) error {
	stmt := tx.StmtContext(tx.ctx, tx.cn.stmt.insertACE)
	defer stmt.Close()

	_, err := stmt.ExecContext(tx.ctx, ent.Path, ent.Perm, principal)
	return normErr(err)
}

// RemoveACEPrincipal implements permission.Transaction.
func (tx *transaction) RemoveACEPrincipal(principal string, ent permission.ACE) error {
	stmt := tx.StmtContext(tx.ctx, tx.cn.stmt.deleteACE)
	defer stmt.Close()

	_, err := stmt.ExecContext(tx.ctx, ent.Path, ent.Perm, principal)
	return normErr(err)
}

// GetAllACEPrincipals implements permission.Transaction.
func (tx *transaction) GetAllACEPrincipals(ents []permission.ACE) ([]string, error) {
	if len(ents) == 0 {
		return nil, nil
	}

	stmt := minisql.Pooled()
	defer minisql.Release(stmt)

	stmt.AppendString("SELECT DISTINCT principal FROM permission_paths WHERE ")
	appendACEConstraints(stmt, ents)

	return scanStringSlice(stmt.QueryContext(tx.ctx, tx))
}

// GetAccessiblePaths implements permission.Transaction.
func (tx *transaction) GetAccessiblePaths(dst []riposo.Path, principals []string, ents []permission.ACE) ([]riposo.Path, error) {
	if len(principals) == 0 || len(ents) == 0 {
		return nil, nil
	}

	stmt := minisql.Pooled()
	defer minisql.Release(stmt)

	stmt.AppendString("SELECT path FROM permission_paths WHERE principal = ANY(")
	stmt.AppendValue(pq.Array(principals))
	stmt.AppendString(") AND ")
	appendACEConstraints(stmt, ents)

	rows, err := stmt.QueryContext(tx.ctx, tx)
	if err != nil {
		return dst, normErr(err)
	}
	defer rows.Close()

	for rows.Next() {
		var path riposo.Path
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}

		dst = append(dst, path)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return dst, nil
}

// GetPermissions implements permission.Transaction.
func (tx *transaction) GetPermissions(path riposo.Path) (schema.PermissionSet, error) {
	stmt := tx.StmtContext(tx.ctx, tx.cn.stmt.getPerms)
	defer stmt.Close()

	rows, err := stmt.QueryContext(tx.ctx, path)
	if err != nil {
		return nil, normErr(err)
	}
	defer rows.Close()

	perms := make(schema.PermissionSet)
	for rows.Next() {
		var perm, principal string
		if err := rows.Scan(&perm, &principal); err != nil {
			return nil, err
		}
		perms[perm] = append(perms[perm], principal)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return perms, nil
}

// CreatePermissions implements permission.Transaction.
func (tx *transaction) CreatePermissions(path riposo.Path, set schema.PermissionSet) error {
	stmt := minisql.Pooled()
	defer minisql.Release(stmt)

	stmt.AppendString("INSERT INTO permission_paths (path, permission, principal) VALUES ")
	first := true
	for perm, principals := range set {
		for _, principal := range principals {
			if first {
				first = false
			} else {
				stmt.AppendByte(',')
			}
			stmt.AppendByte('(')
			stmt.AppendValue(path)
			stmt.AppendByte(',')
			stmt.AppendValue(perm)
			stmt.AppendByte(',')
			stmt.AppendValue(principal)
			stmt.AppendByte(')')
		}
	}
	stmt.AppendString(" ON CONFLICT (path, permission, principal) DO NOTHING")

	_, err := stmt.ExecContext(tx.ctx, tx)
	return normErr(err)
}

// MergePermissions implements permission.Transaction.
func (tx *transaction) MergePermissions(path riposo.Path, set schema.PermissionSet) error {
	if len(set) == 0 {
		return nil
	}

	stmt := minisql.Pooled()
	defer minisql.Release(stmt)

	insert := permsIncludeChanges(set)
	if insert {
		stmt.AppendString("WITH tuples AS (VALUES ")
		first := true
		for perm, principals := range set {
			for _, principal := range principals {
				if first {
					first = false
				} else {
					stmt.AppendByte(',')
				}
				stmt.AppendByte('(')
				stmt.AppendValue(perm)
				stmt.AppendByte(',')
				stmt.AppendValue(principal)
				stmt.AppendByte(')')
			}
		}

		stmt.AppendString("), deletes AS (")
	}

	stmt.AppendString("DELETE FROM permission_paths WHERE path = ")
	stmt.AppendValue(path)
	stmt.AppendString(" AND (")

	first := true
	for perm, principals := range set {
		if first {
			first = false
		} else {
			stmt.AppendString(" OR ")
		}

		stmt.AppendString("permission = ")
		stmt.AppendValue(perm)
		if len(principals) != 0 {
			stmt.AppendString(" AND principal NOT IN (")
			for i, principal := range principals {
				if i != 0 {
					stmt.AppendByte(',')
				}
				stmt.AppendValue(principal)
			}
			stmt.AppendByte(')')
		}
	}
	stmt.AppendByte(')')

	if insert {
		stmt.AppendString(") INSERT INTO permission_paths (path, permission, principal) SELECT ")
		stmt.AppendValue(path)
		stmt.AppendString(", column1, column2 FROM tuples ON CONFLICT (path, permission, principal) DO NOTHING")
	}
	_, err := stmt.ExecContext(tx.ctx, tx)
	return normErr(err)
}

// DeletePermissions implements permission.Transaction.
func (tx *transaction) DeletePermissions(paths []riposo.Path) error {
	if len(paths) == 0 {
		return nil
	}

	stmt := minisql.Pooled()
	defer minisql.Release(stmt)

	// delete exact and nested
	stmt.AppendString("DELETE FROM permission_paths WHERE")
	for i, path := range paths {
		if i != 0 {
			stmt.AppendString(" OR")
		}
		stmt.AppendString(" path = ")
		stmt.AppendValue(path)
		stmt.AppendString(" OR path LIKE ")
		stmt.AppendValue(path + "/%")
	}

	_, err := stmt.ExecContext(tx.ctx, tx)
	return normErr(err)
}

func normErr(err error) error {
	if errors.Is(err, sql.ErrTxDone) {
		return permission.ErrTxDone
	}
	return err
}

func permsIncludeChanges(set schema.PermissionSet) bool {
	for _, principals := range set {
		if len(principals) != 0 {
			return true
		}
	}
	return false
}

func scanStringSlice(rows *sql.Rows, err error) ([]string, error) {
	if err != nil {
		return nil, normErr(err)
	}
	defer rows.Close()

	var res []string
	for rows.Next() {
		var str string
		if err := rows.Scan(&str); err != nil {
			return nil, err
		}
		res = append(res, str)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func appendACEConstraints(stmt *minisql.Query, ents []permission.ACE) {
	stmt.AppendByte('(')
	for i, ent := range ents {
		if i != 0 {
			stmt.AppendString(" OR ")
		}

		stmt.AppendString("(permission = ")
		stmt.AppendValue(ent.Perm)
		stmt.AppendString(" AND path")
		if ent.Path.IsNode() {
			path := strings.TrimSuffix(ent.Path.String(), "*")
			stmt.AppendString(" LIKE ")
			stmt.AppendValue(path + "%")
			stmt.AppendString(" AND path NOT LIKE ")
			stmt.AppendValue(path + "%/%")
		} else {
			stmt.AppendString(" = ")
			stmt.AppendValue(ent.Path)
		}
		stmt.AppendByte(')')
	}
	stmt.AppendByte(')')
}
