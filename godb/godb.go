package godb

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var db *pgx.Conn

var migrationsFS embed.FS

func InitDB() {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	db = conn
}

// ApplyMigrations применяет все миграции из папки migrations
func ApplyMigrations() error {
	// Читаем все файлы из embedded FS
	files, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %v", err)
	}

	// Сортируем файлы по имени (версии)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			content, err := migrationsFS.ReadFile("migrations/" + file.Name())
			if err != nil {
				return fmt.Errorf("failed to read migration %s: %v", file.Name(), err)
			}

			// Выполняем миграцию
			_, err = db.Exec(context.Background(), string(content))
			if err != nil {
				return fmt.Errorf("failed to execute migration %s: %v", file.Name(), err)
			}

			fmt.Printf("Applied migration: %s\n", file.Name())
		}
	}

	return nil
}

// CheckRequestsRemaining проверяет количество оставшихся запросов пользователя
func CheckRequestsRemaining(userID int64) (int, error) {
	var requests int
	err := db.QueryRow(
		context.Background(),
		"SELECT requests_remaining FROM user_balances WHERE user_id = $1",
		userID,
	).Scan(&requests)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("user not found")
	}
	if err != nil {
		return 0, fmt.Errorf("query failed: %w", err)
	}

	return requests, nil
}

// AddNewUser добавляет нового пользователя и инициализирует его баланс
func AddNewUser(userID int64) error {
	tx, err := db.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback(context.Background())

	// Добавляем пользователя
	_, err = tx.Exec(
		context.Background(),
		"INSERT INTO users (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING",
		userID,
	)
	if err != nil {
		return fmt.Errorf("insert user failed: %w", err)
	}

	// Инициализируем баланс
	_, err = tx.Exec(
		context.Background(),
		`INSERT INTO user_balances (user_id, requests_remaining)
         VALUES ($1, 5)
         ON CONFLICT (user_id) DO NOTHING`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("init balance failed: %w", err)
	}

	return tx.Commit(context.Background())
}

// CreateReferralLink создает реферальную ссылку для пользователя
func CreateReferralLink(userID int64) (pgtype.UUID, error) {
	var referralUUID pgtype.UUID
	err := db.QueryRow(
		context.Background(),
		`INSERT INTO referral_links (user_id)
         VALUES ($1)
         RETURNING referral_uuid`,
		userID,
	).Scan(&referralUUID)

	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("failed to create referral link: %w", err)
	}

	return referralUUID, nil
}

// AddRequests добавляет запросы к балансу пользователя
func AddRequests(userID int64, additionalRequests int) error {
	_, err := db.Exec(
		context.Background(),
		`UPDATE user_balances
         SET requests_remaining = requests_remaining + $1
         WHERE user_id = $2`,
		additionalRequests,
		userID,
	)
	if err != nil {
		return fmt.Errorf("update balance failed: %w", err)
	}

	return nil
}
