package middleware

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/jwtauth/v5"

	"github.com/eugene982/url-shortener/internal/logger"
)

const (
	secretKey = "==SuperSecretKey=="
	tokenExp  = time.Hour * 3
)

var (
	tokenAuth  *jwtauth.JWTAuth
	userRandID *rand.Rand
)

type contextKeyType uint

const (
	contextKeyUserID contextKeyType = iota
)

func init() {
	// аутентификация
	tokenAuth = jwtauth.New("HS256", []byte(secretKey), nil)
	userRandID = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func Auth(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {
		var userID int64
		ctx := r.Context()

		_, claims, err := jwtauth.FromContext(ctx)
		// Токен не создат, или истекло время
		if errors.Is(err, jwtauth.ErrNoTokenFound) || errors.Is(err, jwtauth.ErrExpired) {
			userID = userRandID.Int63()

			_, tokenString, err := tokenAuth.Encode(map[string]interface{}{
				"user_id": strconv.FormatInt(userID, 10),
			})

			if err != nil {
				logger.Error(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:    "jwt",
				Value:   tokenString,
				Expires: time.Now().Add(tokenExp),
			})

			ru := r.WithContext(context.WithValue(ctx, contextKeyUserID, userID))
			logger.Info("generate new user id", "user_id", strconv.FormatInt(userID, 10))
			next.ServeHTTP(w, ru)
			return
		}

		// 	любая другая ошибка получения токена
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// токен существует, проверка идентификатора пользователя
		id, ok := claims["user_id"]
		if !ok {
			logger.Warn("user id not found in claims")
			http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
			return
		}

		sid, ok := id.(string)
		if !ok {
			logger.Error(fmt.Errorf("cannot convert to string"), "user_id", id)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		userID, err = strconv.ParseInt(sid, 10, 0)
		if err != nil {
			logger.Error(fmt.Errorf("cannot parse user id string: %w", err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		logger.Info("user is logged", "user_id", userID)
		ru := r.WithContext(context.WithValue(ctx, contextKeyUserID, userID))
		next.ServeHTTP(w, ru)
	}

	return http.HandlerFunc(fn)
}

func Verifier(next http.Handler) http.Handler {
	fn := jwtauth.Verifier(tokenAuth)
	return fn(next)
}

// Возвращает идентификатор пользователя из контекста
func GetUserID(r *http.Request) (int64, error) {
	userID := r.Context().Value(contextKeyUserID)
	if userID == nil {
		return 0, fmt.Errorf("user id not found")
	}
	id, ok := userID.(int64)
	if !ok {
		return 0, fmt.Errorf("user id is not uint type")
	}
	return id, nil
}
