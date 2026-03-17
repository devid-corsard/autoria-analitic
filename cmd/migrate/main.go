package main

import (
	"log"
	"os"
	"path/filepath"

	"personal/autoria/database"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://" + os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") +
			"@" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + "/" + os.Getenv("DB_NAME") + "?sslmode=disable"
	}
	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		dir = "database/migrations"
	}
	absDir, _ := filepath.Abs(dir)
	if err := database.RunMigrations(dsn, absDir); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("migrations ok")
}
