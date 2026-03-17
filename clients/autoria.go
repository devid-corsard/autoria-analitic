package autoria

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const searchBaseURL = "https://developers.ria.com/auto/search"

type Client struct {
	apiKey string
	http   *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey, http: &http.Client{}}
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
}

// SearchResult is the response from the auto search API.
type SearchResult struct {
	Result struct {
		SearchResult struct {
			IDs   []int64 `json:"ids"`
			Count int     `json:"count"`
			LastID int64  `json:"last_id"`
			QS    struct {
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
	q := make(url.Values)
	q.Set("api_key", c.apiKey)
	q.Set("category_id", string(params.CategoryID))
	if params.OrderBy != "" {
		q.Set("order_by", string(params.OrderBy))
	}

	u, err := url.Parse(searchBaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse search url: %w", err)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search api: status %s", resp.Status)
	}
	var out SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &out, nil
}
