package data

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/labstack/echo/v5"
	"ztatic-go-framework/security/crypto"
)

// ErrNotFound is returned when a query expects a row but none is found.
var ErrNotFound = errors.New("ztatic/data: record not found")

// Scanner interface allows abstracting sql.Row and sql.Rows
type Scanner interface {
	Scan(dest ...any) error
}

// BaseRepository provides highly-reusable, generic CRUD operations for any data model.
// Utilizing Go 1.18+ Generics, it eliminates boilerplate SQL rewriting for standard tables.
type BaseRepository[T any] struct {
	DB          *DBEngine
	TableName   string
	CipherSuite *crypto.CipherSuite
}

// NewBaseRepository instantiates a generic repository for the specified SQL table.
func NewBaseRepository[T any](db *DBEngine, tableName string) *BaseRepository[T] {
	return &BaseRepository[T]{
		DB:          db,
		TableName:   tableName,
		CipherSuite: crypto.GetDefaultCipherSuite(),
	}
}

// EncryptModel encrypts fields tagged with `ztatic:"encrypt"` on the provided model.
func (r *BaseRepository[T]) EncryptModel(entity *T) error {
	cs := r.CipherSuite
	if cs == nil {
		cs = crypto.GetDefaultCipherSuite()
	}
	if cs == nil {
		return errors.New("ztatic/data: cipher suite is uninitialized for model encryption; call app.SetCipherKey(...)")
	}
	return crypto.ProcessStruct(entity, cs, true)
}

// DecryptModel decrypts fields tagged with `ztatic:"encrypt"` on the provided model.
func (r *BaseRepository[T]) DecryptModel(entity *T) error {
	cs := r.CipherSuite
	if cs == nil {
		cs = crypto.GetDefaultCipherSuite()
	}
	if cs == nil {
		return errors.New("ztatic/data: cipher suite is uninitialized for model decryption; call app.SetCipherKey(...)")
	}
	return crypto.ProcessStruct(entity, cs, false)
}

// QueryByID retrieves a single entity by its primary key (ID).
// It accepts a custom scanFn to map the SQL columns into the generic Struct T.
// If the entity has fields tagged with `ztatic:"encrypt"`, they are automatically decrypted.
func (r *BaseRepository[T]) QueryByID(ctx context.Context, id any, scanFn func(row Scanner, entity *T) error) (*T, error) {
	query, args, err := r.DB.Builder.
		Select("*").
		From(r.TableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	
	if err != nil {
		return nil, err
	}

	row := r.DB.SQL.QueryRowContext(ctx, query, args...)
	
	var entity T
	if err := scanFn(row, &entity); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	_ = r.DecryptModel(&entity)

	return &entity, nil
}

// DeleteByID removes a record by its primary key.
func (r *BaseRepository[T]) DeleteByID(ctx context.Context, id any) error {
	query, args, err := r.DB.Builder.
		Delete(r.TableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	
	if err != nil {
		return err
	}

	result, err := r.DB.SQL.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		return ErrNotFound
	}
	
	return nil
}

// --- rapid.Resource[T] Interface Default Implementations ---
// These default methods ensure that any repository embedding BaseRepository[T] 
// automatically satisfies the rapid.Resource[T] interface, allowing it to be 
// mounted directly via rapid.RegisterResource. Developers can override these 
// methods in their specific repositories to provide actual implementations.

func (r *BaseRepository[T]) FindAll(c *echo.Context) ([]T, error) {
	return nil, errors.New("ztatic: FindAll not implemented in base repository")
}

func (r *BaseRepository[T]) FindByID(c *echo.Context, id string) (T, error) {
	var empty T
	return empty, errors.New("ztatic: FindByID not implemented in base repository")
}

func (r *BaseRepository[T]) Create(c *echo.Context, item *T) (T, error) {
	var empty T
	return empty, errors.New("ztatic: Create not implemented in base repository")
}

func (r *BaseRepository[T]) Update(c *echo.Context, id string, item *T) (T, error) {
	var empty T
	return empty, errors.New("ztatic: Update not implemented in base repository")
}

func (r *BaseRepository[T]) Delete(c *echo.Context, id string) error {
	return errors.New("ztatic: Delete not implemented in base repository")
}
