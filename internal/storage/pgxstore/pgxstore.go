// Хранение ссылок в базе postgres
package pgxstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/eugene982/url-shortener/internal/logger"
	"github.com/eugene982/url-shortener/internal/model"
	"github.com/eugene982/url-shortener/internal/storage"
)

const (
	maxOpenConns    = 3               // максимум открытых соединений
	maxIdleConns    = 3               // максимум ожидающих соединений
	connMaxLifetime = time.Minute * 3 // таймаут ожидания соединния перед закрытием
)

type PgxStore struct {
	db *sqlx.DB
}

// Утверждение типа, ошибка компиляции
var _ storage.Storage = (*PgxStore)(nil)

// New Функция-конструктор
func New(db *sqlx.DB) (*PgxStore, error) {
	err := db.Ping()
	if err != nil {
		return nil, err
	}

	if err = createTableIfNonExists(db); err != nil {
		return nil, err
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	return &PgxStore{db}, nil
}

// Close Закрытие соединения
func (p *PgxStore) Close() error {
	return p.db.Close()
}

// Ping Пинг к базе
func (p *PgxStore) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// GetAddr Запрос полного адреса у базы по короткой ссылке
func (p *PgxStore) GetAddr(ctx context.Context, short string) (data model.StoreData, err error) {
	query := `
		SELECT * FROM address 
		WHERE short_url=$1 LIMIT 1`

	res := model.StoreData{}
	if err = p.db.GetContext(ctx, &res, query, short); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.StoreData{}, storage.ErrAddressNotFound
		}
		return model.StoreData{}, err
	}
	return res, nil
}

// Set Установка уникального соответствия
func (p *PgxStore) Set(ctx context.Context, data model.StoreData) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			logger.Error(fmt.Errorf("psql rollbacck error: %w", err))
		}
	}()

	query := `
		INSERT INTO address (origin_url, short_url, user_id, is_deleted) 
		VALUES(:origin_url, :short_url, :user_id, :is_deleted);`
	if _, err = tx.NamedExecContext(ctx, query, data); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			err = storage.ErrAddressConflict
		}
		return err
	}
	return tx.Commit()
}

// Update Записть в базу соответствия между адресом и короткой ссылкой
func (p *PgxStore) Update(ctx context.Context, list []model.StoreData) error {
	if len(list) == 0 {
		return nil
	}

	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			logger.Error(fmt.Errorf("psql rollbacck error: %w", err))
		}
	}()

	stmt, err := tx.PrepareNamedContext(ctx, `
		INSERT INTO address 
			(origin_url, short_url, user_id, is_deleted) 
		VALUES
			(:origin_url, :short_url, :user_id, :is_deleted )
		ON CONFLICT (short_url) 
		DO UPDATE SET 
			origin_url=:origin_url, short_url=:short_url,
			user_id=:user_id, is_deleted=:is_deleted;`)
	if err != nil {
		return err
	}

	// Обновляем адреса которые есть в базе и добавляем новые, при отсутствии
	for _, d := range list {
		if _, err = stmt.ExecContext(ctx, d); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetUserURLs Получение данных пользователя
func (p *PgxStore) GetUserURLs(ctx context.Context, userID string) ([]model.StoreData, error) {
	res := make([]model.StoreData, 0)

	query := `
		SELECT * FROM address 
		WHERE user_id=$1`

	err := p.db.SelectContext(ctx, &res, query, userID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// DeleteShort Удаление указанных сокращённых ссылок
func (p *PgxStore) DeleteShort(ctx context.Context, shortURLs []string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	// успокаиваем линтер
	defer func() {
		if err := tx.Rollback(); err != nil {
			logger.Error(fmt.Errorf("psql rollbacck error: %w", err))
		}
	}()

	query, args, err := sqlx.In(`
		UPDATE address SET is_deleted=TRUE  
		WHERE short_url IN (?);`, shortURLs)

	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, p.db.Rebind(query), args...); err != nil {
		return err
	}
	return tx.Commit()
}

// Stats возвращаем статистику
func (p *PgxStore) Stats(ctx context.Context) (URLs int, users int, err error) {
	query := `
		SELECT 
			COUNT(DISTINCT user_id) as Users, 
			COUNT(short_url) as Urls
		FROM address`

	var res struct {
		Users int
		Urls  int
	}

	if err = p.db.GetContext(ctx, &res, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	return res.Urls, res.Users, nil
}

// При первом запуске база может быть пустая
func createTableIfNonExists(db *sqlx.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS address (
			short_url  VARCHAR (20) PRIMARY KEY,
			origin_url TEXT NOT NULL,
			user_id    VARCHAR (36) NOT NULL,
			is_deleted BOOLEAN NOT NULL
		);
		CREATE UNIQUE INDEX IF NOT EXISTS origin_url_idx 
		ON address (origin_url);
		CREATE INDEX IF NOT EXISTS user_id_idx 
		ON address (user_id);`
	_, err := db.Exec(query)
	return err
}
