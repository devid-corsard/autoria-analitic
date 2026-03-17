package autoria

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
)

const (
	searchBaseURL = "https://developers.ria.com/auto/search"
	infoBaseURL   = "https://developers.ria.com/auto/info"
)

type Client struct {
	apiKey string
	resty  *resty.Client
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey, resty: resty.New()}
}

// CategoryID is the RIA category for the search.
type CategoryID string

const (
	CategoryCars         CategoryID = "1"  // Легкові
	CategoryMoto         CategoryID = "2"  // Мото
	CategoryWater        CategoryID = "3"  // Водний транспорт
	CategorySpecial      CategoryID = "4"  // Спецтехніка
	CategoryTrailers     CategoryID = "5"  // Причепи
	CategoryTrucks       CategoryID = "6"  // Вантажівки
	CategoryBuses        CategoryID = "7"  // Автобуси
	CategoryRVs          CategoryID = "8"  // Автобудинки
	CategoryAir          CategoryID = "9"  // Повітряний транспорт
	CategoryAgricultural CategoryID = "10" // Сільгосптехніка
)

// OrderBy is the sort order for search results.
type OrderBy string

const (
	OrderPriceAsc    OrderBy = "2"  // Від дешевих до дорогих
	OrderPriceDesc   OrderBy = "3"  // Від дорогих до дешевих
	OrderNewest      OrderBy = "7"  // Дата додавання новіші
	OrderOldest      OrderBy = "8"  // Дата додавання старіші
	OrderMileageDesc OrderBy = "12" // Пробіг, за спаданням
	OrderMileageAsc  OrderBy = "13" // Пробіг, за зростанням
	OrderYearNewer   OrderBy = "5"  // Рік випуску, новіші авто
	OrderYearOlder   OrderBy = "6"  // Рік випуску, старіші авто
)

// ListParams are parameters for listing cars.
type ListParams struct {
	CategoryID CategoryID
	OrderBy    OrderBy
	// Countpage is page size (max 100). Use "100" for maximum.
	Countpage string
	// Page is 0-based page number for pagination.
	Page string
}

// idsArray unmarshals ids from the search API (can be string or number).
type idsArray []int64

func (a *idsArray) UnmarshalJSON(data []byte) error {
	var raw []interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	out := make([]int64, 0, len(raw))
	for i, v := range raw {
		switch x := v.(type) {
		case string:
			n, err := strconv.ParseInt(x, 10, 64)
			if err != nil {
				return fmt.Errorf("ids[%d]: invalid string %q: %w", i, x, err)
			}
			out = append(out, n)
		case float64:
			out = append(out, int64(x))
		default:
			return fmt.Errorf("ids[%d]: unexpected type %T", i, v)
		}
	}
	*a = out
	return nil
}

// SearchResult is the response from the auto search API.
type SearchResult struct {
	Result struct {
		SearchResult struct {
			IDs    idsArray `json:"ids"`
			Count  int      `json:"count"`
			LastID int64    `json:"last_id"`
			QS     struct {
				Fields []string `json:"fields"`
				Size   int      `json:"size"`
				From   int      `json:"from"`
			} `json:"qs"`
		} `json:"search_result"`
	} `json:"result"`
}

// ListCars runs the auto search with the given list parameters.
// api_key is added automatically.
func (c *Client) ListCars(params ListParams) (*SearchResult, error) {
	q := map[string]string{
		"api_key":     c.apiKey,
		"category_id": string(params.CategoryID),
	}
	if params.OrderBy != "" {
		q["order_by"] = string(params.OrderBy)
	}
	if params.Countpage != "" {
		q["countpage"] = string(params.Countpage)
	}
	if params.Page != "" {
		q["page"] = string(params.Page)
	}

	var out SearchResult
	resp, err := c.resty.R().
		SetQueryParams(q).
		SetResult(&out).
		Get(searchBaseURL)
	if err != nil {
		return nil, fmt.Errorf("list cars (GET search): request: %w", err)
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("list cars (GET search): status %s", resp.Status())
	}
	return &out, nil
}

// AutoInfo is the response from the auto/info API (ad details by id).
// See: https://docs-developers.ria.com/en/used-cars/auto_search_and_info/auto_info
type AutoInfo struct {
	AutoID       int64   `json:"autoId,omitempty"`
	MarkID       int     `json:"markId,omitempty"`
	ModelID      int     `json:"modelId,omitempty"`
	MarkName     string  `json:"markName,omitempty"`
	ModelName    string  `json:"modelName,omitempty"`
	Title        string  `json:"title,omitempty"`
	USD          int     `json:"USD,omitempty"`
	UAH          int     `json:"UAH,omitempty"`
	EUR          int     `json:"EUR,omitempty"`
	Prices       []struct {
		USD string `json:"USD,omitempty"`
		UAH string `json:"UAH,omitempty"`
		EUR string `json:"EUR,omitempty"`
	} `json:"prices,omitempty"`
	LinkToView   string `json:"linkToView,omitempty"`
	VIN          string `json:"VIN,omitempty"`
	AddDate      string `json:"addDate,omitempty"`
	UpdateDate   string `json:"updateDate,omitempty"`
	ExpireDate   string `json:"expireDate,omitempty"`
	LocationCity string `json:"locationCityName,omitempty"`
	AutoData     struct {
		AutoID          int64  `json:"autoId,omitempty"`
		Year            int    `json:"year,omitempty"`
		Race            string `json:"race,omitempty"`
		RaceInt         int    `json:"raceInt,omitempty"`
		Description    string `json:"description,omitempty"`
		FuelName       string `json:"fuelName,omitempty"`
		GearboxName    string `json:"gearboxName,omitempty"`
		CategoryID     int    `json:"categoryId,omitempty"`
		IsSold         bool   `json:"isSold,omitempty"`
		MainCurrency   string `json:"mainCurrency,omitempty"`
		ModificationName string `json:"modificationName,omitempty"`
		GenerationName  string `json:"generationName,omitempty"`
		BodyID         int    `json:"bodyId,omitempty"`
		DriveName      string `json:"driveName,omitempty"`
	} `json:"autoData,omitempty"`
	PhotoData struct {
		All    []int64 `json:"all,omitempty"`
		Count  int     `json:"count,omitempty"`
		SeoLinkM string `json:"seoLinkM,omitempty"`
		SeoLinkB string `json:"seoLinkB,omitempty"`
		SeoLinkF string `json:"seoLinkF,omitempty"`
		SeoLinkSX string `json:"seoLinkSX,omitempty"`
	} `json:"photoData,omitempty"`
	StateData struct {
		StateID   int    `json:"stateId,omitempty"`
		CityID    int    `json:"cityId,omitempty"`
		Name      string `json:"name,omitempty"`
		RegionName string `json:"regionName,omitempty"`
		LinkToCatalog string `json:"linkToCatalog,omitempty"`
	} `json:"stateData,omitempty"`
	Dealer struct {
		ID         int    `json:"id,omitempty"`
		Name       string `json:"name,omitempty"`
		Link       string `json:"link,omitempty"`
		Logo       string `json:"logo,omitempty"`
		Type       string `json:"type,omitempty"`
		Verified   bool   `json:"verified,omitempty"`
		IsReliable bool   `json:"isReliable,omitempty"`
	} `json:"dealer,omitempty"`
	TechnicalCondition struct {
		ID         int    `json:"id,omitempty"`
		Title     string `json:"title,omitempty"`
		Annotation string `json:"annotation,omitempty"`
	} `json:"technicalCondition,omitempty"`
	Color struct {
		Name string `json:"name,omitempty"`
		Eng  string `json:"eng,omitempty"`
		Hex  string `json:"hex,omitempty"`
	} `json:"color,omitempty"`
	ExchangePossible bool `json:"exchangePossible,omitempty"`
	AuctionPossible  bool `json:"auctionPossible,omitempty"`
}

// GetByID returns auto info by announcement id.
// See: https://docs-developers.ria.com/en/used-cars/auto_search_and_info/auto_info
func (c *Client) GetByID(autoID string) (*AutoInfo, error) {
	q := map[string]string{
		"api_key": c.apiKey,
		"auto_id": autoID,
	}
	var out AutoInfo
	resp, err := c.resty.R().
		SetQueryParams(q).
		SetResult(&out).
		Get(infoBaseURL)
	if err != nil {
		return nil, fmt.Errorf("auto info request: %w", err)
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("auto info api: status %s", resp.Status())
	}
	return &out, nil
}
