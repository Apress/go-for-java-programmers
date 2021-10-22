package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq" // can use other drivers
	"strings"
)

// Table row entity
type DBEntity struct {
	name  string
	value string
}

// Do in a DB context.
func DoInDB(driverName, datasourceParams string, f func(db *sql.DB) error) (err error) {
	db, err := sql.Open(driverName, datasourceParams)
	if err != nil {
		return
	}
	defer db.Close()
	err = f(db)
	return
}

// Do in a connection.
func DoInConn(db *sql.DB, ctx context.Context, f func(db *sql.DB, conn *sql.Conn, ctx context.Context) error) (err error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return
	}
	defer conn.Close()
	err = f(db, conn, ctx)
	return
}

// Do in a transaction.
func DoInTx(db *sql.DB, conn *sql.Conn, ctx context.Context, txOptions *sql.TxOptions, f func(tx *sql.Tx) error) (err error) {
	if txOptions == nil {
		txOptions = &sql.TxOptions{Isolation: sql.LevelSerializable}
	}
	tx, err := db.BeginTx(ctx, txOptions)
	if err != nil {
		return
	}
	err = f(tx)
	if err != nil {
		_ = tx.Rollback()
		return
	}
	err = tx.Commit()
	if err != nil {
		return
	}
	return
}

var ErrBadOperation = errors.New("bad operation")

// Execute a SQL statement.
func ExecuteSQL(tx *sql.Tx, ctx context.Context, sql string, params ...interface{}) (count int64, values []*DBEntity, err error) {
	lsql := strings.ToLower(sql)
	switch {

	// process query
	case strings.HasPrefix(lsql, "select "):
		rows, xerr := tx.QueryContext(ctx, sql, params...)
		if xerr != nil {
			err = xerr
			return
		}
		defer rows.Close()
		for rows.Next() {
			var name string
			var value string
			if err = rows.Scan(&name, &value); err != nil {
				return
			}
			data := &DBEntity{name, value}
			values = append(values, data)
		}
		if xerr := rows.Err(); xerr != nil {
			err = xerr
			return
		}

	// process an update
	case strings.HasPrefix(lsql, "update "), strings.HasPrefix(lsql, "delete "), strings.HasPrefix(lsql, "insert "):
		result, xerr := tx.ExecContext(ctx, sql, params...)
		if xerr != nil {
			err = xerr
			return
		}
		count, xerr = result.RowsAffected()
		if xerr != nil {
			err = xerr
			return
		}

	default:
		err = ErrBadOperation // INSERT and DELETE not demo’ed here
		return
	}
	return
}

func testDB() {
	values := make([]*DBEntity, 0, 10)
	values = append(values, &DBEntity{"Barry", "author"},
		&DBEntity{"Barry, Jr.", "reviewer"})

	err := DoInDB("postgres",
		"postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
		func(db *sql.DB) (err error) {
			err = DoInConn(db, context.Background(), func(db *sql.DB, conn *sql.Conn,
				ctx context.Context) (err error) {
				err = createRows(db, conn, ctx, values)
				if err != nil {
					return
				}
				// must be done in separate transaction to see the change
				err = queryRows(db, conn, ctx)
				return
			})
			return
		})
	if err != nil {
		fmt.Printf("DB access failed: %v\n", err)
	}
}

// Create data rows.
func createRows(db *sql.DB, conn *sql.Conn, ctx context.Context, values []*DBEntity) (err error) {
	err = DoInTx(db, conn, ctx, nil, func(tx *sql.Tx) (err error) {
		// first remove any old rows
		count, _, err := ExecuteSQL(tx, ctx, `delete from xvalues`)
		if err != nil {
			return
		}
		fmt.Printf("deleted %d\n", count)
		// insert new rows
		for _, v := range values {
			count1, _, xerr := ExecuteSQL(tx, ctx, fmt.Sprintf(`insert into xvalues(name, value) values('%s', '%s')`, v.name, v.value))
			if xerr != nil || count1 != 1 {
				err = xerr
				return
			}
			fmt.Printf("inserted %q = %q\n", v.name, v.value)
		}
		// update a row
		v := &DBEntity{"Barry", "father"}
		_, _, xerr := ExecuteSQL(tx, ctx, fmt.Sprintf(`update xvalues set value='%s' where name='%s'`, v.value, v.name))
		if xerr != nil {
			err = xerr
			return
		}
		fmt.Printf("updated %q = %q\n", v.name, v.value)
		return
	})
	return
}

// Query and print all rows.
func queryRows(db *sql.DB, conn *sql.Conn, ctx context.Context) (err error) {
	err = DoInTx(db, conn, ctx, nil, func(tx *sql.Tx) (err error) {
		_, xvalues, err := ExecuteSQL(tx, ctx, `select name, value from xvalues`)
		if err != nil {
			return
		}
		for _, v := range xvalues {
			fmt.Printf("queried %q = %q\n", v.name, v.value)
		}
		return
	})
	return
}

func main() {
	testDB()
}
