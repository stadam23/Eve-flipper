package esi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// CharacterOrder represents a character's market order.
type CharacterOrder struct {
	OrderID      int64   `json:"order_id"`
	TypeID       int32   `json:"type_id"`
	LocationID   int64   `json:"location_id"`
	RegionID     int32   `json:"region_id"`
	Price        float64 `json:"price"`
	VolumeRemain int32   `json:"volume_remain"`
	VolumeTotal  int32   `json:"volume_total"`
	IsBuyOrder   bool    `json:"is_buy_order"`
	Duration     int     `json:"duration"`
	Issued       string  `json:"issued"`
	// Enriched fields (filled by server)
	TypeName     string `json:"type_name,omitempty"`
	LocationName string `json:"location_name,omitempty"`
}

// HistoricalOrder represents a completed/cancelled/expired order.
type HistoricalOrder struct {
	OrderID      int64   `json:"order_id"`
	TypeID       int32   `json:"type_id"`
	LocationID   int64   `json:"location_id"`
	RegionID     int32   `json:"region_id"`
	Price        float64 `json:"price"`
	VolumeRemain int32   `json:"volume_remain"`
	VolumeTotal  int32   `json:"volume_total"`
	IsBuyOrder   bool    `json:"is_buy_order"`
	State        string  `json:"state"` // cancelled, expired, fulfilled
	Issued       string  `json:"issued"`
	// Enriched fields
	TypeName     string `json:"type_name,omitempty"`
	LocationName string `json:"location_name,omitempty"`
}

// WalletTransaction represents a wallet transaction.
type WalletTransaction struct {
	TransactionID int64   `json:"transaction_id"`
	Date          string  `json:"date"`
	TypeID        int32   `json:"type_id"`
	LocationID    int64   `json:"location_id"`
	UnitPrice     float64 `json:"unit_price"`
	Quantity      int32   `json:"quantity"`
	IsBuy         bool    `json:"is_buy"`
	// Enriched fields
	TypeName     string `json:"type_name,omitempty"`
	LocationName string `json:"location_name,omitempty"`
}

// SkillEntry represents a single trained skill.
type SkillEntry struct {
	SkillID      int32 `json:"skill_id"`
	ActiveLevel  int   `json:"active_skill_level"`
	TrainedLevel int   `json:"trained_skill_level"`
	SkillPoints  int64 `json:"skillpoints_in_skill"`
}

// SkillSheet is the character's skill data.
type SkillSheet struct {
	Skills    []SkillEntry `json:"skills"`
	TotalSP   int64        `json:"total_sp"`
	UnallocSP int64        `json:"unallocated_sp"`
}

// CharacterLocation represents the character's current location.
type CharacterLocation struct {
	SolarSystemID int32 `json:"solar_system_id"`
	StationID     int64 `json:"station_id,omitempty"`
	StructureID   int64 `json:"structure_id,omitempty"`
}

// --- Authenticated requests using the shared Client transport ---

// AuthGetJSON performs an authenticated GET to an ESI endpoint using the shared HTTP
// transport (connection pooling, keep-alive). Uses the lightweight semaphore so that
// character API calls never compete with bulk scan page fetches.
func (c *Client) AuthGetJSON(url, accessToken string, dst interface{}) error {
	c.sem <- struct{}{} // acquire lightweight semaphore

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		<-c.sem
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "eve-flipper/1.0 (github.com)")

	resp, err := c.http.Do(req)
	if err != nil {
		<-c.sem
		return err
	}

	statusCode := resp.StatusCode
	if statusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		<-c.sem
		return fmt.Errorf("ESI %d: %s", statusCode, string(body))
	}

	decErr := json.NewDecoder(resp.Body).Decode(dst)
	resp.Body.Close()
	<-c.sem
	return decErr
}

// GetCharacterOrders fetches a character's active market orders.
func (c *Client) GetCharacterOrders(characterID int64, accessToken string) ([]CharacterOrder, error) {
	url := fmt.Sprintf("%s/characters/%d/orders/?datasource=tranquility", baseURL, characterID)
	var orders []CharacterOrder
	if err := c.AuthGetJSON(url, accessToken, &orders); err != nil {
		return nil, fmt.Errorf("character orders: %w", err)
	}
	return orders, nil
}

// GetWalletBalance fetches a character's ISK balance.
func (c *Client) GetWalletBalance(characterID int64, accessToken string) (float64, error) {
	url := fmt.Sprintf("%s/characters/%d/wallet/?datasource=tranquility", baseURL, characterID)
	var balance float64
	if err := c.AuthGetJSON(url, accessToken, &balance); err != nil {
		return 0, fmt.Errorf("wallet: %w", err)
	}
	return balance, nil
}

// GetSkills fetches a character's trained skills.
func (c *Client) GetSkills(characterID int64, accessToken string) (*SkillSheet, error) {
	url := fmt.Sprintf("%s/characters/%d/skills/?datasource=tranquility", baseURL, characterID)
	var sheet SkillSheet
	if err := c.AuthGetJSON(url, accessToken, &sheet); err != nil {
		return nil, fmt.Errorf("skills: %w", err)
	}
	return &sheet, nil
}

// GetOrderHistory fetches all pages of a character's completed/cancelled/expired orders.
// ESI may return multiple pages via X-Pages header; this fetches them all concurrently.
func (c *Client) GetOrderHistory(characterID int64, accessToken string) ([]HistoricalOrder, error) {
	historyURL := fmt.Sprintf("%s/characters/%d/orders/history/?datasource=tranquility", baseURL, characterID)

	// Fetch page 1 to discover total pages.
	c.sem <- struct{}{}

	req, err := http.NewRequest("GET", historyURL+"&page=1", nil)
	if err != nil {
		<-c.sem
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "eve-flipper/1.0 (github.com)")

	resp, err := c.http.Do(req)
	if err != nil {
		<-c.sem
		return nil, fmt.Errorf("order history page 1: %w", err)
	}

	totalPages := 1
	if p := resp.Header.Get("X-Pages"); p != "" {
		if tp, parseErr := strconv.Atoi(p); parseErr == nil && tp > 1 {
			totalPages = tp
		}
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		<-c.sem
		return nil, fmt.Errorf("order history: ESI %d: %s", resp.StatusCode, string(body))
	}

	var page1 []HistoricalOrder
	decErr := json.NewDecoder(resp.Body).Decode(&page1)
	resp.Body.Close()
	<-c.sem

	if decErr != nil {
		return nil, fmt.Errorf("order history decode: %w", decErr)
	}

	if totalPages <= 1 {
		return page1, nil
	}

	// Fetch remaining pages concurrently.
	type pageResult struct {
		data []HistoricalOrder
		err  error
	}
	results := make(chan pageResult, totalPages-1)

	for p := 2; p <= totalPages; p++ {
		go func(pageNum int) {
			pageURL := fmt.Sprintf("%s&page=%d", historyURL, pageNum)
			var data []HistoricalOrder
			if fetchErr := c.AuthGetJSON(pageURL, accessToken, &data); fetchErr != nil {
				results <- pageResult{err: fetchErr}
				return
			}
			results <- pageResult{data: data}
		}(p)
	}

	all := make([]HistoricalOrder, 0, len(page1)*totalPages)
	all = append(all, page1...)
	for i := 0; i < totalPages-1; i++ {
		r := <-results
		if r.err != nil {
			continue // skip failed pages, return what we have
		}
		all = append(all, r.data...)
	}
	return all, nil
}

// GetWalletTransactions fetches a character's wallet transactions.
func (c *Client) GetWalletTransactions(characterID int64, accessToken string) ([]WalletTransaction, error) {
	url := fmt.Sprintf("%s/characters/%d/wallet/transactions/?datasource=tranquility", baseURL, characterID)
	var txns []WalletTransaction
	if err := c.AuthGetJSON(url, accessToken, &txns); err != nil {
		return nil, fmt.Errorf("wallet transactions: %w", err)
	}
	return txns, nil
}

// GetCharacterLocation fetches a character's current location (system/station).
func (c *Client) GetCharacterLocation(characterID int64, accessToken string) (*CharacterLocation, error) {
	url := fmt.Sprintf("%s/characters/%d/location/?datasource=tranquility", baseURL, characterID)
	var loc CharacterLocation
	if err := c.AuthGetJSON(url, accessToken, &loc); err != nil {
		return nil, fmt.Errorf("location: %w", err)
	}
	return &loc, nil
}
