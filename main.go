package main

import (
	"context"
	"fmt"
	"os"
	autoria "personal/autoria/clients"
	"personal/autoria/database"
	"personal/autoria/transform"
	"strconv"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const maxIDs = 1000   // cap so we don't exceed GetByID request limit
const countpage = 100 // max page size for search API

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
func LoadConfig(log *zap.Logger) Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Warn("loading .env failed", zap.Error(err))
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
	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	cfg := LoadConfig(log)
	ctx := context.Background()

	db, err := database.Open(ctx, cfg.DSN())
	if err != nil {
		log.Fatal("database open failed", zap.Error(err))
	}
	defer db.DB.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatal("database ping failed", zap.Error(err))
	}

	client := autoria.NewClient(cfg.APIKey)

	// Fetch 100 ids per page, save each page to DB, until we've saved maxIDs total.
	params := autoria.ListParams{
		CategoryID: autoria.CategoryCars,
		OrderBy:    autoria.OrderNewest,
		Countpage:  strconv.Itoa(countpage),
	}
	saved := 0
	for page := 0; saved < maxIDs; page++ {
		params.Page = strconv.Itoa(page)
		result, err := client.ListCars(params)
		if err != nil {
			log.Error("list cars failed", zap.Error(err), zap.Int("page", page))
			break
		}
		log.Info("list cars page", zap.Int("page", page), zap.Int("total_count", result.Result.SearchResult.Count), zap.Int("ids_on_page", len(result.Result.SearchResult.IDs)))
		pageIDs := []int64(result.Result.SearchResult.IDs)
		if len(pageIDs) == 0 {
			break
		}
		toSave := pageIDs
		if left := maxIDs - saved; len(toSave) > left {
			toSave = toSave[:left]
		}
		if err := db.InsertIDs(ctx, toSave); err != nil {
			log.Fatal("insert ids failed", zap.Error(err))
		}
		saved += len(toSave)
		if len(pageIDs) < countpage || saved >= maxIDs {
			break
		}
	}

	// Fetch all ids from DB with no details yet and get details for each.
	ids, err := db.GetIDsPendingDetails(ctx)
	if err != nil {
		log.Fatal("get ids pending details failed", zap.Error(err))
	}
	log.Info("fetching details for cars", zap.Int("count", len(ids)))
	for _, id := range ids {
		info, err := client.GetByID(strconv.FormatInt(id, 10))
		if err != nil {
			log.Error("get by id failed", zap.Int64("id", id), zap.Error(err))
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
			log.Error("update car failed", zap.Int64("id", id), zap.Error(err))
			continue
		}
	}
}
