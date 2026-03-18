package app

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	autoria "personal/autoria/clients"
	"personal/autoria/database"
	"personal/autoria/transform"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	csvColLink      = 0
	csvColTitleYear = 4
	csvColPriceUSD  = 5
	csvColMileage   = 7
)

var idFromLinkRE = regexp.MustCompile(`(\d+)\.html`)

const maxIDs = 1000   // cap so we don't exceed GetByID request limit
const countpage = 100 // max page size for search API

type App struct {
	autoria *autoria.Client
	db      *database.DB
	log     *zap.Logger
}

func New(autoria *autoria.Client, db *database.DB, log *zap.Logger) *App {
	return &App{autoria, db, log}
}

// fetchNewIDs fetches up to maxIDs car IDs from the API and inserts them into the DB.
func (a *App) FetchNewIDs(ctx context.Context) {
	params := autoria.ListParams{
		CategoryID: autoria.CategoryCars,
		OrderBy:    autoria.OrderNewest,
		Countpage:  strconv.Itoa(countpage),
	}
	saved := 0
	for page := 0; saved < maxIDs; page++ {
		params.Page = strconv.Itoa(page)
		result, err := a.autoria.ListCars(params)
		if err != nil {
			a.log.Error("list cars failed", zap.Error(err), zap.Int("page", page))
			break
		}
		a.log.Info("list cars page", zap.Int("page", page), zap.Int("total_count", result.Result.SearchResult.Count), zap.Int("ids_on_page", len(result.Result.SearchResult.IDs)))
		pageIDs := []int64(result.Result.SearchResult.IDs)
		if len(pageIDs) == 0 {
			break
		}
		toSave := pageIDs
		if left := maxIDs - saved; len(toSave) > left {
			toSave = toSave[:left]
		}
		if err := a.db.InsertIDs(ctx, toSave); err != nil {
			a.log.Fatal("insert ids failed", zap.Error(err))
		}
		saved += len(toSave)
		if len(pageIDs) < countpage || saved >= maxIDs {
			break
		}
	}
}

// fillEmptyDetails fetches details from the API for all DB records that have no details yet and updates them.
func (a *App) FillEmptyDetails(ctx context.Context) error {
	ids, err := a.db.GetIDsPendingDetails(ctx)
	if err != nil {
		return fmt.Errorf("get ids pending details failed: %w", err)
	}
	a.log.Info("fetching details for cars", zap.Int("count", len(ids)))
	for _, id := range ids {
		info, err := a.autoria.GetByID(strconv.FormatInt(id, 10))
		if err != nil {
			if strings.Contains(err.Error(), "429") {
				return fmt.Errorf("rate limited (429), stopping: %w", err)
			}
			a.log.Error("get by id failed", zap.Int64("id", id), zap.Error(err))
			continue
		}
		car := transform.AutoInfoToCar(info)
		if car == nil {
			continue
		}
		if car.ID == 0 {
			car.ID = id
		}
		if err := a.db.Update(ctx, car); err != nil {
			a.log.Error("update car failed", zap.Int64("id", id), zap.Error(err))
			continue
		}
	}
	return nil
}

// extractIDFromLink extracts the numeric auto ID from a RIA link (e.g. ..._39610235.html). Returns 0, false if not found.
func extractIDFromLink(link string) (int64, bool) {
	m := idFromLinkRE.FindStringSubmatch(link)
	if len(m) < 2 {
		return 0, false
	}
	id, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

// parseTitleYear splits "Make Model YYYY" into title and year; year is the last token if it's 4 digits.
func parseTitleYear(s string) (title string, year int) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", 0
	}
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return s, 0
	}
	last := parts[len(parts)-1]
	if len(last) == 4 {
		if y, err := strconv.Atoi(last); err == nil && y >= 1900 && y <= 2100 {
			title = strings.TrimSpace(strings.Join(parts[:len(parts)-1], " "))
			return title, y
		}
	}
	return s, 0
}

// parseUSD parses price like "77 000 $" (spaces and nbsp), returns 0 on error.
func parseUSD(s string) int {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\u00a0", "")
	s = strings.TrimSuffix(s, "$")
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n, _ := strconv.Atoi(s)
	return n
}

// parseMileage parses "65 тис. км" to km (65000). Removes " тис. км" and multiplies by 1000.
func parseMileage(s string) int {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, " тис. км")
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n * 1000
}

// LoadCSV reads a CSV file by path and inserts normalized rows into the cars table.
// CSV columns: link href (0), common-text 3 (4)=title+year, common-text 5 (6)=usd, common-text 7 (8)=mileage.
// Invalid rows (e.g. missing ID) are skipped with a log warning.
func (a *App) LoadCSV(ctx context.Context, csvPath string) error {
	f, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("open csv: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("read csv: %w", err)
	}

	if len(records) < 2 {
		a.log.Info("csv has no data rows")
		return nil
	}

	// skip header
	rows := records[1:]
	now := time.Now()
	created := 0
	skipped := 0

	for i, rec := range rows {
		if len(rec) <= csvColMileage {
			skipped++
			a.log.Debug("row too short", zap.Int("row", i+2), zap.Int("cols", len(rec)))
			continue
		}

		id, ok := extractIDFromLink(rec[csvColLink])
		if !ok {
			skipped++
			a.log.Warn("skip row: no id in link", zap.Int("row", i+2), zap.String("link", rec[csvColLink]))
			continue
		}

		title, year := parseTitleYear(rec[csvColTitleYear])
		usd := parseUSD(rec[csvColPriceUSD])
		raceInt := parseMileage(rec[csvColMileage])

		car := &database.Car{
			ID:         id,
			Title:      title,
			Year:       year,
			USD:        usd,
			RaceInt:    raceInt,
			LinkToView: strings.TrimSpace(rec[csvColLink]),
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		if err := a.db.Upsert(ctx, car); err != nil {
			a.log.Error("upsert car failed", zap.Int64("id", id), zap.Error(err))
			skipped++
			continue
		}
		created++
	}

	a.log.Info("load csv done", zap.String("path", csvPath), zap.Int("created", created), zap.Int("skipped", skipped))
	return nil
}
