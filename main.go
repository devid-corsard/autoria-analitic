package main

import (
	"context"
	"flag"
	"math/rand"
	"personal/autoria/app"
	autoria "personal/autoria/clients"
	"personal/autoria/config"
	"personal/autoria/database"
	"time"

	"go.uber.org/zap"
)

func main() {
	fetchNew := flag.Bool("fetch-new", false, "fetch new car IDs from API and save to DB; otherwise only fill details for empty records")
	loadCSV := flag.String("load", "", "load cars from CSV file and exit (e.g. -load ./auto.csv)")
	flag.Parse()

	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	cfg := config.LoadConfig(log)
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
	a := app.New(client, db, log)

	if *loadCSV != "" {
		if err := a.LoadCSV(ctx, *loadCSV); err != nil {
			log.Fatal("load csv failed", zap.Error(err))
		}
		return
	}

	if *fetchNew {
		a.FetchNewIDs(ctx)
	}

	for {
		if err := a.FillEmptyDetails(ctx); err != nil {
			timeTOWait := time.Duration(rand.Intn(20)+20) * time.Minute
			log.Info("waiting before retrying", zap.Float64("minutes", timeTOWait.Minutes()))
			time.Sleep(timeTOWait)
			continue
		}
		break
	}
}
