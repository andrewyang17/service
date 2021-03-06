// Package database provides support for access the database.
package database

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/andrewyang17/service/foundation/web"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

var (
	ErrDBNotFound        = errors.New("not found")
	ErrDBDuplicatedEntry = errors.New("duplicated entry")
)

type Config struct {
	User         string
	Password     string
	Host         string
	Name         string
	MaxIdleConns int
	MaxOpenConns int
	DisableTLS   bool
}

func Open(cfg Config) (*sqlx.DB, error) {
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	db, err := sqlx.Open("postgres", u.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)

	return db, nil
}

func StatusCheck(ctx context.Context, db *sqlx.DB) error {
	var pingError error
	for attempts := 1; ; attempts++ {
		pingError = db.Ping()
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Run a simple query to determine connectivity. Running this query forces a
	// round trip through the database.
	const q = `SELECT TRUE`
	var tmp bool

	return db.QueryRowContext(ctx, q).Scan(&tmp)
}

type Transactor interface {
	Beginx() (*sqlx.Tx, error)
}

// WithinTran runs passed function and do commit/rollback at the end.
func WithinTran(ctx context.Context, log *zap.SugaredLogger, db Transactor, fn func(extContext sqlx.ExtContext) error) error {
	traceID := web.GetTraceID(ctx)

	log.Infow("begin tran", "traceID", traceID)
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tran: %w", err)
	}

	mustRollBack := true
	defer func() {
		if mustRollBack {
			log.Infow("rollback tran", "traceID", traceID)
			if err := tx.Rollback(); err != nil {
				log.Errorw("unable to rollback tran", "traceID", traceID)
			}
		}
	}()

	if err := fn(tx); err != nil {
		return fmt.Errorf("exec tran: %w", err)
	}

	mustRollBack = false

	log.Infow("commit tran", "traceID", traceID)
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tran: %w", err)
	}

	return nil
}

// NamedExecContext is a helper function to execute a CUD operation with
// logging and tracing.
func NamedExecContext(ctx context.Context, log *zap.SugaredLogger, db sqlx.ExtContext, query string, data interface{}) error {
	q := queryString(query, data)
	log.Infow("database.NameExecContext", "traceID", web.GetTraceID(ctx), "query", q)

	if _, err := sqlx.NamedExecContext(ctx, db, query, data); err != nil {
		return err
	}

	return nil
}

// NamedQuerySlice is a helper function for executing queries that return a
// collection of data to be unmarshalled into a slice.
func NamedQuerySlice(ctx context.Context, log *zap.SugaredLogger, db sqlx.ExtContext, query string, data interface{}, dest interface{}) error {
	q := queryString(query, data)
	log.Infow("database.NamedQuerySlice", "traceID", web.GetTraceID(ctx), "query", q)

	val := reflect.ValueOf(dest)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		return errors.New("must provide a pointer to a slice")
	}

	rows, err := sqlx.NamedQueryContext(ctx, db, query, data)
	if err != nil {
		return err
	}
	defer rows.Close()

	slice := val.Elem()
	for rows.Next() {
		v := reflect.New(slice.Type().Elem())
		if err := rows.StructScan(v.Interface()); err != nil {
			return err
		}
		slice.Set(reflect.Append(slice, v.Elem()))
	}

	return nil
}

// NamedQueryStruct is a helper function for executing queries that return a
// single value to be unmarshalled into a struct type
func NamedQueryStruct(ctx context.Context, log *zap.SugaredLogger, db sqlx.ExtContext, query string, data interface{}, dest interface{}) error {
	q := queryString(query, data)
	log.Infow("databases.NamedQueryStruct", "traceID", web.GetTraceID(ctx), "query", q)

	rows, err := sqlx.NamedQueryContext(ctx, db, query, data)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return ErrDBNotFound
	}

	if err := rows.StructScan(dest); err != nil {
		return err
	}

	return nil
}

// queryString provides a pretty print version of the query and parameters.
func queryString(query string, args ...interface{}) string {
	query, params, err := sqlx.Named(query, args)
	if err != nil {
		return err.Error()
	}

	for _, param := range params {
		var value string
		switch v := param.(type) {
		case string:
			value = fmt.Sprintf("%q", v)
		case []byte:
			value = fmt.Sprintf("%q", string(v))
		default:
			value = fmt.Sprintf("%v", v)
		}
		query = strings.Replace(query, "?", value, 1)
	}

	query = strings.ReplaceAll(query, "\t", "")
	query = strings.ReplaceAll(query, "\n", " ")

	return strings.Trim(query, " ")
}
