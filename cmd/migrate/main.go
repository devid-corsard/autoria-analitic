package main

import (
	"os"
	"path/filepath"

	"personal/autoria/database"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

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
		log.Fatal("migrate failed", zap.Error(err))
	}
	log.Info("migrations completed")
}
