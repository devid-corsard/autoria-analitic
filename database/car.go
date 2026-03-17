package database

import "time"

// Car is the domain model for the cars table (RIA auto_id as primary key).
type Car struct {
	ID                     int64
	MarkID                 int
	ModelID                int
	MarkName               string
	ModelName              string
	Title                  string
	USD                    int
	UAH                    int
	EUR                    int
	LinkToView             string
	VIN                    string
	AddDate                string
	UpdateDate             string
	ExpireDate             string
	LocationCity           string
	Year                   int
	RaceInt                int
	Description            string
	FuelName               string
	GearboxName            string
	CategoryID             int
	IsSold                 bool
	StateID                int
	CityID                 int
	RegionName             string
	DealerID               int
	DealerName             string
	TechnicalConditionID   int
	TechnicalConditionName string
	ColorName              string
	ExchangePossible       bool
	AuctionPossible        bool
	CreatedAt              time.Time
	UpdatedAt              time.Time
}
