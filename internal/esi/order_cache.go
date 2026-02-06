package esi

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// orderCacheKey identifies a cached set of region orders.
type orderCacheKey struct {
	RegionID  int32
	OrderType string // "sell" or "buy"
}

// orderCacheEntry holds cached orders together with HTTP caching metadata.
type orderCacheEntry struct {
	orders  []MarketOrder
	etag    string    // ETag from ESI response (page 1)
	expires time.Time // parsed Expires header
}

// OrderCache is a thread-safe in-memory cache for region market orders.
// It uses ETag/Expires headers from ESI to avoid re-downloading unchanged data.
// A singleflight.Group prevents duplicate in-flight fetches for the same key.
type OrderCache struct {
	mu      sync.RWMutex
	entries map[orderCacheKey]*orderCacheEntry
	group   singleflight.Group
}

// NewOrderCache creates an empty order cache.
func NewOrderCache() *OrderCache {
	return &OrderCache{
		entries: make(map[orderCacheKey]*orderCacheEntry),
	}
}

// Get returns cached orders if they exist and have not expired.
// Returns (orders, etag, hit).
func (oc *OrderCache) Get(regionID int32, orderType string) ([]MarketOrder, string, bool) {
	oc.mu.RLock()
	defer oc.mu.RUnlock()

	e, ok := oc.entries[orderCacheKey{regionID, orderType}]
	if !ok {
		return nil, "", false
	}
	if time.Now().After(e.expires) {
		// Expired — return etag for conditional request, but signal miss.
		return nil, e.etag, false
	}
	return e.orders, e.etag, true
}

// Put stores orders in the cache with the given etag and expiry.
func (oc *OrderCache) Put(regionID int32, orderType string, orders []MarketOrder, etag string, expires time.Time) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	oc.entries[orderCacheKey{regionID, orderType}] = &orderCacheEntry{
		orders:  orders,
		etag:    etag,
		expires: expires,
	}
}

// Touch updates the expiry of an existing cache entry (used on 304 Not Modified).
func (oc *OrderCache) Touch(regionID int32, orderType string, expires time.Time) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	key := orderCacheKey{regionID, orderType}
	if e, ok := oc.entries[key]; ok {
		e.expires = expires
	}
}

// FetchRegionOrdersCached fetches region orders with full caching support:
//  1. If orders are in cache and not expired → instant return
//  2. If orders expired but we have an ETag → conditional request (If-None-Match)
//     - 304: touch expiry, return cached data (no body transfer)
//     - 200: full re-fetch, update cache
//  3. Cache miss → full fetch, populate cache
//
// Uses singleflight to coalesce concurrent requests for the same region+orderType.
func (c *Client) FetchRegionOrdersCached(regionID int32, orderType string) ([]MarketOrder, error) {
	sfKey := fmt.Sprintf("%d:%s", regionID, orderType)

	result, err, _ := c.orderCache.group.Do(sfKey, func() (interface{}, error) {
		return c.fetchRegionOrdersWithCache(regionID, orderType)
	})
	if err != nil {
		return nil, err
	}
	return result.([]MarketOrder), nil
}

// fetchRegionOrdersWithCache is the actual implementation behind singleflight.
func (c *Client) fetchRegionOrdersWithCache(regionID int32, orderType string) ([]MarketOrder, error) {
	// 1. Check cache
	orders, etag, hit := c.orderCache.Get(regionID, orderType)
	if hit {
		log.Printf("[ESI] OrderCache HIT region=%d type=%s (%d orders)", regionID, orderType, len(orders))
		return orders, nil
	}

	url := fmt.Sprintf("%s/markets/%d/orders/?datasource=tranquility&order_type=%s",
		baseURL, regionID, orderType)

	// 2. If we have an ETag, try conditional request on page 1
	if etag != "" {
		notModified, newExpires, err := c.conditionalCheck(url+"&page=1", etag)
		if err == nil && notModified {
			// 304 — data unchanged, refresh expiry
			c.orderCache.Touch(regionID, orderType, newExpires)
			cached, _, _ := c.orderCache.Get(regionID, orderType)
			if cached != nil {
				log.Printf("[ESI] OrderCache 304 region=%d type=%s (ETag match)", regionID, orderType)
				return cached, nil
			}
		}
		// ETag miss or error — fall through to full fetch
	}

	// 3. Full fetch
	allOrders, respEtag, respExpires, err := c.getPaginatedDirectWithHeaders(url, regionID)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.orderCache.Put(regionID, orderType, allOrders, respEtag, respExpires)
	log.Printf("[ESI] OrderCache MISS region=%d type=%s (%d orders, expires=%s)",
		regionID, orderType, len(allOrders), respExpires.Format("15:04:05"))

	return allOrders, nil
}

// conditionalCheck sends a HEAD-like GET with If-None-Match.
// Returns (notModified, newExpires, error).
func (c *Client) conditionalCheck(pageURL, etag string) (bool, time.Time, error) {
	c.scanSem <- struct{}{}
	defer func() { <-c.scanSem }()

	req, err := newESIRequest(pageURL)
	if err != nil {
		return false, time.Time{}, err
	}
	req.Header.Set("If-None-Match", etag)

	resp, err := c.http.Do(req)
	if err != nil {
		return false, time.Time{}, err
	}
	resp.Body.Close()

	expires := parseExpires(resp)

	if resp.StatusCode == 304 {
		return true, expires, nil
	}
	return false, expires, nil
}

// newESIRequest creates a standard ESI GET request with common headers.
func newESIRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "eve-flipper/1.0 (github.com)")
	req.Header.Set("Accept", "application/json")
	return req, nil
}

// parseExpires reads the Expires header from an ESI response.
// Falls back to 5-minute TTL if header is missing or unparseable.
func parseExpires(resp *http.Response) time.Time {
	if exp := resp.Header.Get("Expires"); exp != "" {
		if t, err := time.Parse(time.RFC1123, exp); err == nil {
			return t
		}
	}
	// Fallback: ESI market orders typically refresh every 5 minutes.
	return time.Now().Add(5 * time.Minute)
}
