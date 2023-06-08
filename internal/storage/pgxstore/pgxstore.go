// Хранение ссылок в базе postgres
package pgxstore

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/eugene982/url-shortener/internal/storage"
)

type PgxStore struct {
	db *sql.DB
}

// Утверждение типа, ошибка компиляции
var _ storage.Storage = (*PgxStore)(nil)

// Функция конструктор
func New(databaseDSN string) (*PgxStore, error) {
	//"postgres://username:password@localhost:5432/database_name"
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	if err = createTable(db); err != nil {
		return nil, err
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(3)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(3 * time.Minute)

	return &PgxStore{db}, nil
}

// Закрытие соединения
func (p *PgxStore) Close() error {
	return p.db.Close()
}

// Пинг к базе
func (p *PgxStore) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// Запрос полного адреса у базы по короткой ссылке
func (p *PgxStore) GetAddr(ctx context.Context, short string) (addr string, err error) {
	query := `SELECT addr FROM address WHERE short=$1`

	err = p.db.QueryRowContext(ctx, query, short).Scan(&addr)
	if errors.Is(sql.ErrNoRows, err) {
		return "", storage.ErrAddressNotFound
	}
	return addr, err
}

// Записть в базу соответствия между адресом и короткой ссылкой
func (p *PgxStore) Set(ctx context.Context, addr string, short string) error {
	var query string

	_, err := p.GetAddr(ctx, short)
	if err == nil {
		query = `UPDATE address SET addr=$1 WHERE short=$2`
	} else if err == storage.ErrAddressNotFound {
		query = `INSERT INTO address (addr, short) VALUES($1, $2)`
	} else {
		return err
	}

	_, err = p.db.ExecContext(ctx, query, addr, short)
	return err
}

// При первом запуске база может быть пустая
func createTable(db *sql.DB) error {

	query :=
		`CREATE TABLE IF NOT EXISTS address (
		short VARCHAR (20) PRIMARY KEY,
		addr TEXT NOT NULL
	)`
	_, err := db.Exec(query)
	return err
}
