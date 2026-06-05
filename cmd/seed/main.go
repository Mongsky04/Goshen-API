package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	godotenv.Load() //nolint:errcheck

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	email := "admin@goshen.id"
	password := "admin123"
	if len(os.Args) >= 3 {
		email = os.Args[1]
		password = os.Args[2]
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("bcrypt: %v", err)
	}

	_, err = pool.Exec(context.Background(),
		`INSERT INTO admins (email, password_hash) VALUES ($1, $2)
		 ON CONFLICT (email) DO UPDATE SET password_hash = $2`,
		email, string(hash),
	)
	if err != nil {
		log.Fatalf("insert admin: %v", err)
	}

	fmt.Printf("✓ Admin created/updated: %s\n", email)
}
