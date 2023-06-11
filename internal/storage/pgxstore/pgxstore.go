// Хранение ссылок в базе postgres
package pgxstore

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/eugene982/url-shortener/internal/model"
	"github.com/eugene982/url-shortener/internal/storage"
)

type PgxStore struct {
	db *sqlx.DB
}

// Утверждение типа, ошибка компиляции
var _ storage.Storage = (*PgxStore)(nil)

// Функция конструктор
func New(databaseDSN string) (*PgxStore, error) {
	//"postgres://username:password@localhost:5432/database_name"
	db, err := sqlx.Open("pgx", databaseDSN)
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
	query := `
		SELECT addr FROM address 
		WHERE short=$1 LIMIT 1`

	res := make([]string, 0, 1)
	if err = p.db.SelectContext(ctx, &res, query, short); err != nil {
		return "", err
	}

	if len(res) == 0 {
		return "", storage.ErrAddressNotFound
	} else {
		return res[0], nil
	}
}

// Записть в базу соответствия между адресом и короткой ссылкой
func (p *PgxStore) Set(ctx context.Context, data ...model.StoreData) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if len(data) == 0 {
		return nil
	}

	shorts := make([]string, len(data)) // массив коротких адресов которые поищем в базе
	for i, d := range data {
		shorts[i] = d.ShortURL
	}

	// Поиск в базе уже установленных адресов по списку
	query, args, err := sqlx.In(`
		SELECT addr, short FROM address 
	 	WHERE short IN(?)`, shorts)
	if err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, tx.Rebind(query), args...)
	if err != nil {
		return err
	}

	update := make(map[string]string) // адреса которые нужно перезаписать
	for rows.Next() {
		var short, addr string
		if err = rows.Scan(&addr, &short); err != nil {
			if errors.Is(sql.ErrNoRows, err) {
				continue
			}
			return err
		}
		update[short] = addr
	}
	if rows.Err() != nil {
		return rows.Err()
	}

	// Обновляем адреса которые есть в базе и добавляем новые, при отсутствии
	for _, d := range data {
		if _, ok := update[d.ShortURL]; ok {
			query = `UPDATE address SET addr=$1 WHERE short=$2`
		} else {
			query = `INSERT INTO address (addr, short) VALUES($1, $2)`
		}

		if _, err = tx.ExecContext(ctx, query, d.OriginalURL, d.ShortURL); err != nil {
			return err
		}

	}

	return tx.Commit()
}

// При первом запуске база может быть пустая
func createTable(db *sqlx.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS address (
			short VARCHAR (20) PRIMARY KEY,
			addr TEXT NOT NULL
		);
		CREATE UNIQUE INDEX IF NOT EXISTS short_idx 
		ON address (short);`
	_, err := db.Exec(query)
	return err
}
