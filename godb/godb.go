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
func AddRequests(userID int64, requestsToAdd int) error {
	ctx := context.Background()

	_, err := pool.Exec(ctx, `
		UPDATE users
		SET requests_remaining = requests_remaining + $1
		WHERE user_id = $2
	`, requestsToAdd, userID)

	return err
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
