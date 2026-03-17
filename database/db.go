package database

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// DB wraps *sql.DB and provides car operations.
type DB struct {
	*sql.DB
}

// Open opens a Postgres connection and returns DB. Caller should call Ping to verify.
func Open(ctx context.Context, dsn string) (*DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql open: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return &DB{DB: db}, nil
}

// Ping verifies the database connection is alive. Panic or return error as desired.
func (db *DB) Ping(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}

// RunMigrations runs all pending up migrations from the database/migrations directory.
// migrationsDir is the path to the migrations folder (e.g. "database/migrations" when run from module root).
func RunMigrations(dsn string, migrationsDir string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("postgres driver: %w", err)
	}
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("abs path: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+absPath, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

// InsertIDs inserts one row per ID with only id and created_at. Uses ON CONFLICT DO NOTHING.
func (db *DB) InsertIDs(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO cars (id, created_at, updated_at) VALUES ($1, NOW(), NOW()) ON CONFLICT (id) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, id := range ids {
		_, err := stmt.ExecContext(ctx, id)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Create inserts a full car row. Use InsertIDs + Update for the main flow; Create for optional full insert.
func (db *DB) Create(ctx context.Context, car *Car) error {
	q := `INSERT INTO cars (
		id, mark_id, model_id, mark_name, model_name, title, usd, uah, eur,
		link_to_view, vin, add_date, update_date, expire_date, location_city,
		year, race_int, description, fuel_name, gearbox_name, category_id, is_sold,
		state_id, city_id, region_name, dealer_id, dealer_name,
		technical_condition_id, technical_condition_name, color_name,
		exchange_possible, auction_possible, created_at, updated_at
	) VALUES (
		$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,
		$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34
	) ON CONFLICT (id) DO NOTHING`
	_, err := db.ExecContext(ctx, q,
		car.ID, car.MarkID, car.ModelID, car.MarkName, car.ModelName, car.Title,
		car.USD, car.UAH, car.EUR, car.LinkToView, car.VIN, car.AddDate, car.UpdateDate, car.ExpireDate, car.LocationCity,
		car.Year, car.RaceInt, car.Description, car.FuelName, car.GearboxName, car.CategoryID, car.IsSold,
		car.StateID, car.CityID, car.RegionName, car.DealerID, car.DealerName,
		car.TechnicalConditionID, car.TechnicalConditionName, car.ColorName,
		car.ExchangePossible, car.AuctionPossible, car.CreatedAt, car.UpdatedAt,
	)
	return err
}

// Update updates a car by id. Sets updated_at to now.
func (db *DB) Update(ctx context.Context, car *Car) error {
	q := `UPDATE cars SET
		mark_id=$2, model_id=$3, mark_name=$4, model_name=$5, title=$6, usd=$7, uah=$8, eur=$9,
		link_to_view=$10, vin=$11, add_date=$12, update_date=$13, expire_date=$14, location_city=$15,
		year=$16, race_int=$17, description=$18, fuel_name=$19, gearbox_name=$20, category_id=$21, is_sold=$22,
		state_id=$23, city_id=$24, region_name=$25, dealer_id=$26, dealer_name=$27,
		technical_condition_id=$28, technical_condition_name=$29, color_name=$30,
		exchange_possible=$31, auction_possible=$32, updated_at=NOW()
		WHERE id=$1`
	_, err := db.ExecContext(ctx, q,
		car.ID, car.MarkID, car.ModelID, car.MarkName, car.ModelName, car.Title,
		car.USD, car.UAH, car.EUR, car.LinkToView, car.VIN, car.AddDate, car.UpdateDate, car.ExpireDate, car.LocationCity,
		car.Year, car.RaceInt, car.Description, car.FuelName, car.GearboxName, car.CategoryID, car.IsSold,
		car.StateID, car.CityID, car.RegionName, car.DealerID, car.DealerName,
		car.TechnicalConditionID, car.TechnicalConditionName, car.ColorName,
		car.ExchangePossible, car.AuctionPossible,
	)
	return err
}

// Get returns a car by id.
func (db *DB) Get(ctx context.Context, id int64) (*Car, error) {
	q := `SELECT id, mark_id, model_id, mark_name, model_name, title, usd, uah, eur,
		link_to_view, vin, add_date, update_date, expire_date, location_city,
		year, race_int, description, fuel_name, gearbox_name, category_id, is_sold,
		state_id, city_id, region_name, dealer_id, dealer_name,
		technical_condition_id, technical_condition_name, color_name,
		exchange_possible, auction_possible, created_at, updated_at
		FROM cars WHERE id=$1`
	var c Car
	err := db.QueryRowContext(ctx, q, id).Scan(
		&c.ID, &c.MarkID, &c.ModelID, &c.MarkName, &c.ModelName, &c.Title,
		&c.USD, &c.UAH, &c.EUR, &c.LinkToView, &c.VIN, &c.AddDate, &c.UpdateDate, &c.ExpireDate, &c.LocationCity,
		&c.Year, &c.RaceInt, &c.Description, &c.FuelName, &c.GearboxName, &c.CategoryID, &c.IsSold,
		&c.StateID, &c.CityID, &c.RegionName, &c.DealerID, &c.DealerName,
		&c.TechnicalConditionID, &c.TechnicalConditionName, &c.ColorName,
		&c.ExchangePossible, &c.AuctionPossible, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// GetIDsPendingDetails returns ids of cars that have no details yet (only id/created_at from InsertIDs).
func (db *DB) GetIDsPendingDetails(ctx context.Context) ([]int64, error) {
	rows, err := db.QueryContext(ctx, `SELECT id FROM cars WHERE title = '' ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
