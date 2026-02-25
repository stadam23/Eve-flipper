package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"eve-flipper/internal/engine"
	"eve-flipper/internal/esi"
)

func resolveContractTypeName(sdeName, esiName string, typeID int32) string {
	live := strings.TrimSpace(esiName)
	if live != "" {
		// Prefer live ESI name: contract item naming can drift vs local SDE.
		return live
	}
	local := strings.TrimSpace(sdeName)
	if local != "" {
		return local
	}
	return fmt.Sprintf("Type %d", typeID)
}

// ContractItemResponse represents a contract item in the API response
type ContractItemResponse struct {
	TypeID             int32   `json:"type_id"`
	TypeName           string  `json:"type_name"`
	Quantity           int32   `json:"quantity"`
	IsIncluded         bool    `json:"is_included"`
	IsBlueprintCopy    bool    `json:"is_blueprint_copy"`
	GroupID            int32   `json:"group_id,omitempty"`
	GroupName          string  `json:"group_name,omitempty"`
	CategoryID         int32   `json:"category_id,omitempty"`
	IsShip             bool    `json:"is_ship,omitempty"`
	IsRig              bool    `json:"is_rig,omitempty"`
	RecordID           int64   `json:"record_id"`
	ItemID             int64   `json:"item_id"`
	MaterialEfficiency int     `json:"material_efficiency,omitempty"`
	TimeEfficiency     int     `json:"time_efficiency,omitempty"`
	Runs               int     `json:"runs,omitempty"`
	Flag               int     `json:"flag,omitempty"`
	Singleton          bool    `json:"singleton,omitempty"`
	Damage             float64 `json:"damage,omitempty"`
}

// ContractDetailsResponse represents contract details with items
type ContractDetailsResponse struct {
	ContractID int32                  `json:"contract_id"`
	Items      []ContractItemResponse `json:"items"`
}

// handleGetContractItems returns the items for a specific contract
// GET /api/contracts/{contract_id}/items
func (s *Server) handleGetContractItems(w http.ResponseWriter, r *http.Request) {
	// Extract contract_id from path
	contractIDStr := r.PathValue("contract_id")
	contractID, err := strconv.ParseInt(contractIDStr, 10, 32)
	if err != nil {
		http.Error(w, `{"error":"invalid_contract_id"}`, http.StatusBadRequest)
		return
	}

	// Prefer scanner-level contract-items cache to avoid repeated ESI calls.
	var items []esi.ContractItem
	if s.scanner != nil && s.scanner.ContractItemsCache != nil {
		batch := s.esi.FetchContractItemsBatch(
			[]int32{int32(contractID)},
			s.scanner.ContractItemsCache,
			func(done, total int) {},
		)
		if cached, ok := batch[int32(contractID)]; ok {
			items = cached
		}
	}
	// Fallback direct fetch if cache path had no entry (e.g. transient ESI errors).
	if len(items) == 0 {
		fetched, err := s.esi.FetchContractItems(int32(contractID))
		if err != nil {
			log.Printf("[API] FetchContractItems error: contract_id=%d, err=%v", contractID, err)
			http.Error(w, `{"error":"esi_error"}`, http.StatusInternalServerError)
			return
		}
		items = fetched
	}

	// Convert to response format with type names
	responseItems := make([]ContractItemResponse, 0, len(items))
	liveTypeNames := make(map[int32]string)
	esiTypeName := func(typeID int32) string {
		if v, ok := liveTypeNames[typeID]; ok {
			return v
		}
		name := ""
		if s.esi != nil {
			name = strings.TrimSpace(s.esi.TypeName(typeID))
		}
		liveTypeNames[typeID] = name
		return name
	}

	for _, item := range items {
		if item.Quantity > 0 && engine.IsMarketDisabledTypeID(item.TypeID) {
			continue
		}
		typeNameSDE := ""
		groupID := int32(0)
		groupName := ""
		categoryID := int32(0)
		isShip := false
		isRig := false
		if s.sdeData != nil {
			if t, ok := s.sdeData.Types[item.TypeID]; ok {
				typeNameSDE = t.Name
				groupID = t.GroupID
				categoryID = t.CategoryID
				isShip = t.CategoryID == 6
				isRig = t.IsRig
				if g, ok := s.sdeData.Groups[t.GroupID]; ok {
					groupName = g.Name
					isRig = g.IsRig
				}
			}
		}
		typeName := resolveContractTypeName(typeNameSDE, esiTypeName(item.TypeID), item.TypeID)

		resp := ContractItemResponse{
			TypeID:          item.TypeID,
			TypeName:        typeName,
			Quantity:        item.Quantity,
			IsIncluded:      item.IsIncluded,
			IsBlueprintCopy: item.IsBlueprintCopy,
			GroupID:         groupID,
			GroupName:       groupName,
			CategoryID:      categoryID,
			IsShip:          isShip,
			IsRig:           isRig,
			RecordID:        item.RecordID,
			ItemID:          item.ItemID,
		}

		// Only include blueprint fields if relevant
		if item.IsBlueprintCopy || item.MaterialEfficiency > 0 || item.TimeEfficiency > 0 || item.Runs > 0 {
			resp.MaterialEfficiency = item.MaterialEfficiency
			resp.TimeEfficiency = item.TimeEfficiency
			resp.Runs = item.Runs
		}
		if item.Flag != 0 {
			resp.Flag = item.Flag
		}
		if item.Singleton {
			resp.Singleton = true
		}
		if item.Damage > 0 {
			resp.Damage = item.Damage
		}

		responseItems = append(responseItems, resp)
	}

	response := ContractDetailsResponse{
		ContractID: int32(contractID),
		Items:      responseItems,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
