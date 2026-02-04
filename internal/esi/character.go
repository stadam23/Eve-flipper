package esi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
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
	SkillID       int32 `json:"skill_id"`
	ActiveLevel   int   `json:"active_skill_level"`
	TrainedLevel  int   `json:"trained_skill_level"`
	SkillPoints   int64 `json:"skillpoints_in_skill"`
}

// SkillSheet is the character's skill data.
type SkillSheet struct {
	Skills     []SkillEntry `json:"skills"`
	TotalSP    int64        `json:"total_sp"`
	UnallocSP  int64        `json:"unallocated_sp"`
}

// GetCharacterOrders fetches a character's active market orders.
func GetCharacterOrders(characterID int64, accessToken string) ([]CharacterOrder, error) {
	url := fmt.Sprintf("%s/characters/%d/orders/?datasource=tranquility", baseURL, characterID)
	var orders []CharacterOrder
	if err := authGet(url, accessToken, &orders); err != nil {
		return nil, fmt.Errorf("character orders: %w", err)
	}
	return orders, nil
}

// GetWalletBalance fetches a character's ISK balance.
func GetWalletBalance(characterID int64, accessToken string) (float64, error) {
	url := fmt.Sprintf("%s/characters/%d/wallet/?datasource=tranquility", baseURL, characterID)
	var balance float64
	if err := authGet(url, accessToken, &balance); err != nil {
		return 0, fmt.Errorf("wallet: %w", err)
	}
	return balance, nil
}

// GetSkills fetches a character's trained skills.
func GetSkills(characterID int64, accessToken string) (*SkillSheet, error) {
	url := fmt.Sprintf("%s/characters/%d/skills/?datasource=tranquility", baseURL, characterID)
	var sheet SkillSheet
	if err := authGet(url, accessToken, &sheet); err != nil {
		return nil, fmt.Errorf("skills: %w", err)
	}
	return &sheet, nil
}

// GetOrderHistory fetches a character's completed/cancelled/expired orders.
func GetOrderHistory(characterID int64, accessToken string) ([]HistoricalOrder, error) {
	url := fmt.Sprintf("%s/characters/%d/orders/history/?datasource=tranquility", baseURL, characterID)
	var orders []HistoricalOrder
	if err := authGet(url, accessToken, &orders); err != nil {
		return nil, fmt.Errorf("order history: %w", err)
	}
	return orders, nil
}

// GetWalletTransactions fetches a character's wallet transactions.
func GetWalletTransactions(characterID int64, accessToken string) ([]WalletTransaction, error) {
	url := fmt.Sprintf("%s/characters/%d/wallet/transactions/?datasource=tranquility", baseURL, characterID)
	var txns []WalletTransaction
	if err := authGet(url, accessToken, &txns); err != nil {
		return nil, fmt.Errorf("wallet transactions: %w", err)
	}
	return txns, nil
}

// CharacterLocation represents the character's current location.
type CharacterLocation struct {
	SolarSystemID int32 `json:"solar_system_id"`
	StationID     int64 `json:"station_id,omitempty"`
	StructureID   int64 `json:"structure_id,omitempty"`
}

// GetCharacterLocation fetches a character's current location (system/station).
func GetCharacterLocation(characterID int64, accessToken string) (*CharacterLocation, error) {
	url := fmt.Sprintf("%s/characters/%d/location/?datasource=tranquility", baseURL, characterID)
	var loc CharacterLocation
	if err := authGet(url, accessToken, &loc); err != nil {
		return nil, fmt.Errorf("location: %w", err)
	}
	return &loc, nil
}

// authGet performs an authenticated GET request to an ESI endpoint.
func authGet(url, accessToken string, dst interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "eve-flipper/1.0 (github.com)")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ESI %d: %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(dst)
}
