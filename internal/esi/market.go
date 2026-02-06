package esi

import (
	"fmt"
)

// MarketOrder mirrors the ESI market order response.
type MarketOrder struct {
	OrderID      int64   `json:"order_id"`
	TypeID       int32   `json:"type_id"`
	LocationID   int64   `json:"location_id"`
	SystemID     int32   `json:"system_id"`
	Price        float64 `json:"price"`
	VolumeRemain int32   `json:"volume_remain"`
	IsBuyOrder   bool    `json:"is_buy_order"`
	RegionID     int32   `json:"-"` // set by us
}

// FetchRegionOrders fetches all market orders for a region.
// Uses in-memory cache with ETag/Expires — repeated calls within the ESI refresh
// window (typically 5 min) return instantly without any network I/O.
func (c *Client) FetchRegionOrders(regionID int32, orderType string) ([]MarketOrder, error) {
	return c.FetchRegionOrdersCached(regionID, orderType)
}

// FetchRegionOrdersByType fetches all market orders for a specific type in a region.
func (c *Client) FetchRegionOrdersByType(regionID int32, typeID int32) ([]MarketOrder, error) {
	url := fmt.Sprintf("%s/markets/%d/orders/?datasource=tranquility&order_type=all&type_id=%d",
		baseURL, regionID, typeID)

	return c.GetPaginatedDirect(url, regionID)
}
