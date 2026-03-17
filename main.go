package main

import (
	"fmt"
	"log"
	"os"
	"personal/autoria/clients"
)

func main() {
	apiKey := os.Getenv("api_key")
	if apiKey == "" {
		log.Fatal("api_key environment variable is required")
	}
	client := autoria.NewClient(apiKey)

	params := autoria.ListParams{
		CategoryID: autoria.CategoryCars,
		OrderBy:    autoria.OrderNewest,
	}
	result, err := client.ListCars(params)
	if err != nil {
		log.Fatalf("list cars: %v", err)
	}
	sr := &result.Result.SearchResult
	fmt.Printf("count: %d, ids (first 10): %v\n", sr.Count, firstN(sr.IDs, 10))
}

func firstN(a []int64, n int) []int64 {
	if len(a) <= n {
		return a
	}
	return a[:n]
}
