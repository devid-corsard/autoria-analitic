package transform

import (
	"strconv"
	"time"

	"personal/autoria/clients"
	"personal/autoria/database"
)

// AutoInfoToCar maps autoria AutoInfo to the database Car model.
func AutoInfoToCar(info *autoria.AutoInfo) *database.Car {
	if info == nil {
		return nil
	}
	now := time.Now()
	c := &database.Car{
		CreatedAt: now,
		UpdatedAt: now,
	}
	// ID from AutoData.autoId or top-level
	if info.AutoData.AutoID != 0 {
		c.ID = info.AutoData.AutoID
	} else if info.AutoID != 0 {
		c.ID = info.AutoID
	}
	c.MarkID = info.MarkID
	c.ModelID = info.ModelID
	c.MarkName = info.MarkName
	c.ModelName = info.ModelName
	c.Title = info.Title
	c.USD = info.USD
	c.UAH = info.UAH
	c.EUR = info.EUR
	c.LinkToView = info.LinkToView
	c.VIN = info.VIN
	c.AddDate = info.AddDate
	c.UpdateDate = info.UpdateDate
	c.ExpireDate = info.ExpireDate
	c.LocationCity = info.LocationCity
	c.Year = info.AutoData.Year
	c.RaceInt = info.AutoData.RaceInt
	c.Description = info.AutoData.Description
	c.FuelName = info.AutoData.FuelName
	c.GearboxName = info.AutoData.GearboxName
	c.CategoryID = info.AutoData.CategoryID
	c.IsSold = info.AutoData.IsSold
	c.StateID = info.StateData.StateID
	c.CityID = info.StateData.CityID
	c.RegionName = info.StateData.RegionName
	c.DealerID = info.Dealer.ID
	c.DealerName = info.Dealer.Name
	c.TechnicalConditionID = info.TechnicalCondition.ID
	c.TechnicalConditionName = info.TechnicalCondition.Title
	c.ColorName = info.Color.Name
	c.ExchangePossible = info.ExchangePossible
	c.AuctionPossible = info.AuctionPossible
	return c
}

// MustParseAutoID converts id to string for GetByID. No-op helper.
func MustParseAutoID(id int64) string {
	return strconv.FormatInt(id, 10)
}
