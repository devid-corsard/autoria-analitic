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
			log.Printf("error listing cars: %v", err)
			break
		}
		log.Printf("total results: %v\n", result.Result.SearchResult.Count)
		pageIDs := []int64(result.Result.SearchResult.IDs)
		if len(pageIDs) == 0 {
			break
		}
		toSave := pageIDs
		if left := maxIDs - saved; len(toSave) > left {
			toSave = toSave[:left]
		}
		if err := db.InsertIDs(ctx, toSave); err != nil {
			log.Fatalf("%v", err)
		}
		saved += len(toSave)
		if len(pageIDs) < countpage || saved >= maxIDs {
			break
		}
	}

	// Fetch all ids from DB with no details yet and get details for each.
	ids, err := db.GetIDsPendingDetails(ctx)
	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("total ids to fetch details for: %v", len(ids))
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
