package engine

// marketDisabledTypeIDs lists item types that may appear in ESI market data
// but are not practically tradable via normal sell-side execution.
// Keep this list conservative: only hard-verified market-disabled types.
var marketDisabledTypeIDs = map[int32]struct{}{
	MPTCTypeID: {}, // Multiple Pilot Training Certificate
}

const playerStructureLocationIDMin int64 = 1_000_000_000_000

func isMarketDisabledType(typeID int32) bool {
	_, blocked := marketDisabledTypeIDs[typeID]
	return blocked
}

// isPlayerStructureLocationID reports whether a market location id belongs to an Upwell structure.
func isPlayerStructureLocationID(locationID int64) bool {
	return locationID > playerStructureLocationIDMin
}

// IsMarketDisabledTypeID reports whether the given type is known market-disabled.
// Exported for API-level safety filters.
func IsMarketDisabledTypeID(typeID int32) bool {
	return isMarketDisabledType(typeID)
}

// IsPlayerStructureLocationID is exported for API-level safety filters.
func IsPlayerStructureLocationID(locationID int64) bool {
	return isPlayerStructureLocationID(locationID)
}
