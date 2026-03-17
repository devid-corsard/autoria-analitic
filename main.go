package main

import (
	"context"
	"fmt"
	"log"
	"os"
	autoria "personal/autoria/clients"
	"personal/autoria/database"
	"personal/autoria/transform"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds app configuration from .env.
type Config struct {
	APIKey     string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

// DSN returns the Postgres connection string.
func (c Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

// LoadConfig loads .env and returns Config. Panics if required vars are missing.
func LoadConfig() Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("warning: loading .env: %v", err)
	}
	cfg := Config{
		APIKey:     os.Getenv("api_key"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
	}
	if cfg.APIKey == "" {
		panic("api_key is required")
	}
	if cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBUser == "" || cfg.DBPassword == "" || cfg.DBName == "" {
		panic("DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME are required")
	}
	return cfg
}

func main() {
	log.SetFlags(log.Lshortfile)

	cfg := LoadConfig()
	ctx := context.Background()

	db, err := database.Open(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer db.DB.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("%v", err)
	}

	client := autoria.NewClient(cfg.APIKey)

	params := autoria.ListParams{
		CategoryID: autoria.CategoryCars,
		OrderBy:    autoria.OrderNewest,
	}
	result, err := client.ListCars(params)
	if err != nil {
		log.Fatalf("%v", err)
	}
	ids := []int64(result.Result.SearchResult.IDs)

	if err := db.InsertIDs(ctx, ids); err != nil {
		log.Fatalf("%v", err)
	}

	for _, id := range ids {
		info, err := client.GetByID(strconv.FormatInt(id, 10))
		if err != nil {
			log.Printf("%v", err)
			continue
		}
		car := transform.AutoInfoToCar(info)
		if car == nil {
			continue
		}
		if car.ID == 0 {
			car.ID = id
		}
		if err := db.Update(ctx, car); err != nil {
			log.Printf("%v", err)
			continue
		}
	}
}
