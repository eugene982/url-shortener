// Хранение ссылок в базе postgres
package pgxstore

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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

// Установка уникального соответствия
func (p *PgxStore) Set(ctx context.Context, addr, short string) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO address (addr, short) VALUES($1, $2)`
	if _, err = tx.ExecContext(ctx, query, addr, short); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			err = storage.ErrAddressConflict
		}
		return err
	}
	return tx.Commit()
}

// Записть в базу соответствия между адресом и короткой ссылкой
func (p *PgxStore) Update(ctx context.Context, data []model.StoreData) error {
	if len(data) == 0 {
		return nil
	}

	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO address (addr, short) VALUES($1, $2)
		ON CONFLICT (short) 
		DO UPDATE SET addr=$1, short=$2`)
	if err != nil {
		return err
	}

	// Обновляем адреса которые есть в базе и добавляем новые, при отсутствии
	for _, d := range data {
		if _, err = stmt.ExecContext(ctx, d.OriginalURL, d.ShortURL); err != nil {
			return err
		}
	}

	return tx.Commit()
}
