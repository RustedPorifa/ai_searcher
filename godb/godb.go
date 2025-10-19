package godb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

type User struct {
	UserID             int64   `json:"user_id"`
	Role               string  `json:"role"`
	RequestsRemaining  int     `json:"requests_remaining"`
	ReferredBy         *string `json:"referred_by,omitempty"`
	IsReferralRewarded bool    `json:"is_referral_rewarded"`
}

func InitDB() {
	dbURL := os.Getenv("DATABASE_URL")
	// Контекст для создания пула соединений
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Создание пула соединений
	var poolErr error
	pool, poolErr = pgxpool.New(ctx, dbURL)
	if poolErr != nil {
		log.Panic("Ошибка в создании пула соединения: ", poolErr)
	}

	//Проверка подключения к базе данных
	if pingErr := pool.Ping(ctx); pingErr != nil {
		log.Panic("Ошибка проверки подключения к базе данных: ", pingErr)
	}

	//Запуск миграции баз данных
	runMigration()
}

// Запускает мигррацию баз данных
func runMigration() {
	migrationSQL, readErr := os.ReadDir("./godb/migration")
	if readErr != nil {
		log.Panic("Ошибка чтения директории миграций: ", readErr)
	}
	for _, file := range migrationSQL {
		println(file.Name())
		filePath := "./godb/migration/" + file.Name()
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			log.Panic("Ошибка чтения файла миграции: ", err)
		}
		_, err = pool.Exec(context.Background(), string(fileContent))
		if err != nil {
			log.Panic("Ошибка выполнения миграции: ", err)
		}
	}
}

// AddAdminsIfNotExist добавляет пользователей как администраторов, если они не существуют
// Если пользователь уже существует, обновляет его роль на администратора
func AddAdminsIfNotExist(userIDs []int64) error {
	ctx := context.Background()
	if pool == nil {
		return errors.New("pool is not initialized")
	}

	if len(userIDs) == 0 {
		return nil
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Сначала обновляем существующих пользователей до администраторов
	_, err = tx.Exec(ctx, `
        UPDATE users
        SET role = 'admin'
        WHERE user_id = ANY($1)
    `, userIDs)
	if err != nil {
		return fmt.Errorf("failed to update existing users to admin: %w", err)
	}

	// Затем добавляем новых пользователей как администраторов
	for _, userID := range userIDs {
		// Вставляем пользователя с ролью администратора
		_, err = tx.Exec(ctx, `
            INSERT INTO users (user_id, role)
            VALUES ($1, 'admin')
            ON CONFLICT (user_id) DO NOTHING
        `, userID)
		if err != nil {
			return fmt.Errorf("failed to insert admin user %d: %w", userID, err)
		}

		// Создаем реферальную ссылку для нового пользователя
		_, err = tx.Exec(ctx, `
            INSERT INTO referral_links (user_id)
            VALUES ($1)
            ON CONFLICT (user_id) DO NOTHING
        `, userID)
		if err != nil {
			return fmt.Errorf("failed to create referral link for admin %d: %w", userID, err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}
	tx = nil
	return nil
}

// AddUserIfNotExists добавляет пользователя если его нет и создает реферальную ссылку
// referralUUID - UUID пользователя, который пригласил (может быть пустой строкой)
func AddUserIfNotExists(userID int64, referralUUID string) error {
	ctx := context.Background()
	if pool == nil {
		return errors.New("pool is not initialized")
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	// Безопасный defer
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Валидируем и проверяем существование referralUUID
	var referredBy *string
	if referralUUID != "" {
		var exists bool
		err = tx.QueryRow(ctx, `
			SELECT true
			FROM referral_links
			WHERE referral_uuid = $1::uuid
		`, referralUUID).Scan(&exists)
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("check referral failed: %w", err)
		}
		if err == nil { // найден
			referredBy = &referralUUID
		}
		// если не найден — referredBy остаётся nil
	}

	// Вставляем пользователя
	_, err = tx.Exec(ctx, `
		INSERT INTO users (user_id, referred_by)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO NOTHING
	`, userID, referredBy)
	if err != nil {
		return fmt.Errorf("insert user failed: %w", err)
	}

	// Создаём реф.ссылку
	_, err = tx.Exec(ctx, `
		INSERT INTO referral_links (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING
	`, userID)
	if err != nil {
		return fmt.Errorf("insert referral link failed: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}
	tx = nil // необязательно, но явно
	return nil
}

// GetUserIDByReferralLink ищет user_id по реферальной ссылке (uuid)
func GetUserIDByReferralLink(referralUUID string) (int64, error) {
	ctx := context.Background()

	var userID int64
	err := pool.QueryRow(ctx, `
		SELECT user_id
		FROM referral_links
		WHERE referral_uuid = $1
	`, referralUUID).Scan(&userID)

	if err != nil {
		return 0, err
	}

	return userID, nil
}

// AddRequests добавляет запросы пользователю по переданному количеству
// Только администраторы могут выполнять эту операцию
func AddRequests(adminUserID, targetUserID int64, requestsToAdd int) error {
	ctx := context.Background()

	// Сначала проверяем, является ли пользователь администратором
	var role string
	err := pool.QueryRow(ctx, `
        SELECT role
        FROM users
        WHERE user_id = $1
    `, adminUserID).Scan(&role)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("admin user not found")
		}
		return fmt.Errorf("failed to verify admin role: %w", err)
	}

	if role != "admin" {
		return fmt.Errorf("user is not in admin group")
	}

	// Выполняем добавление запросов целевому пользователю
	result, err := pool.Exec(ctx, `
        UPDATE users
        SET requests_remaining = requests_remaining + $1
        WHERE user_id = $2
    `, requestsToAdd, targetUserID)

	if err != nil {
		return fmt.Errorf("failed to add requests: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("target user not found")
	}

	return nil
}

// GetUserReferralLink возвращает referral_uuid для заданного пользователя
func GetUserReferralLink(userID int64) (string, error) {
	ctx := context.Background()

	var referralUUID string
	err := pool.QueryRow(ctx, `
		SELECT referral_uuid
		FROM referral_links
		WHERE user_id = $1
	`, userID).Scan(&referralUUID)

	if err != nil {
		return "", err
	}

	return referralUUID, nil
}

// GetUser возвращает полную информацию о пользователе по его user_id
func GetUser(userID int64) (*User, error) {
	ctx := context.Background()

	var user User
	var referredByUUID *string

	err := pool.QueryRow(ctx, `
		SELECT user_id, role, requests_remaining, referred_by, is_refferal_rewarded
		FROM users
		WHERE user_id = $1
	`, userID).Scan(
		&user.UserID,
		&user.Role,
		&user.RequestsRemaining,
		&referredByUUID,
		&user.IsReferralRewarded,
	)

	if err != nil {
		return nil, err
	}

	user.ReferredBy = referredByUUID
	return &user, nil
}

// DecrementUserRequests уменьшает количество запросов пользователя на 1, если есть доступные запросы
// Возвращает новое количество запросов или ошибку
func DecrementUserRequests(userID int64) (int, error) {
	ctx := context.Background()

	var newRequests int
	err := pool.QueryRow(ctx, `
        UPDATE users
        SET requests_remaining = requests_remaining - 1
        WHERE user_id = $1 AND requests_remaining > 0
        RETURNING requests_remaining
    `, userID).Scan(&newRequests)

	if err != nil {
		if err == pgx.ErrNoRows {
			// Проверяем, существует ли пользователь и сколько у него запросов
			var currentRequests int
			err := pool.QueryRow(ctx, `
                SELECT requests_remaining
                FROM users
                WHERE user_id = $1
            `, userID).Scan(&currentRequests)

			if err != nil {
				if err == pgx.ErrNoRows {
					return 0, fmt.Errorf("user not found")
				}
				return 0, fmt.Errorf("failed to check user requests: %w", err)
			}

			if currentRequests <= 0 {
				return 0, fmt.Errorf("no requests available")
			}

			return 0, fmt.Errorf("unexpected error: failed to decrement requests")
		}
		return 0, fmt.Errorf("failed to decrement user requests: %w", err)
	}

	return newRequests, nil
}

// IncrementUserRequests увеличивает количество оставшихся запросов пользователя на 1.
// Возвращает новое количество запросов или ошибку.
func IncrementUserRequests(userID int64) (int, error) {
	ctx := context.Background()

	var newRequests int
	err := pool.QueryRow(ctx, `
        UPDATE users
        SET requests_remaining = requests_remaining + 1
        WHERE user_id = $1
        RETURNING requests_remaining
    `, userID).Scan(&newRequests)

	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("user not found")
		}
		return 0, fmt.Errorf("failed to increment user requests: %w", err)
	}

	return newRequests, nil
}

// GetUserRequests возвращает количество оставшихся запросов пользователя
func GetUserRequests(userID int64) (int, error) {
	ctx := context.Background()

	var requestsRemaining int
	err := pool.QueryRow(ctx, `
        SELECT requests_remaining
        FROM users
        WHERE user_id = $1
    `, userID).Scan(&requestsRemaining)

	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("user not found")
		}
		return 0, fmt.Errorf("failed to get user requests: %w", err)
	}

	return requestsRemaining, nil
}
