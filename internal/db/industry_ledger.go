package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	IndustryProjectStatusDraft     = "draft"
	IndustryProjectStatusPlanned   = "planned"
	IndustryProjectStatusActive    = "active"
	IndustryProjectStatusCompleted = "completed"
	IndustryProjectStatusArchived  = "archived"

	IndustryTaskStatusPlanned   = "planned"
	IndustryTaskStatusReady     = "ready"
	IndustryTaskStatusActive    = "active"
	IndustryTaskStatusPaused    = "paused"
	IndustryTaskStatusCompleted = "completed"
	IndustryTaskStatusBlocked   = "blocked"
	IndustryTaskStatusCancelled = "cancelled"

	IndustryJobStatusPlanned   = "planned"
	IndustryJobStatusQueued    = "queued"
	IndustryJobStatusActive    = "active"
	IndustryJobStatusPaused    = "paused"
	IndustryJobStatusCompleted = "completed"
	IndustryJobStatusFailed    = "failed"
	IndustryJobStatusCancelled = "cancelled"
)

type IndustryProject struct {
	ID        int64           `json:"id"`
	UserID    string          `json:"user_id"`
	Name      string          `json:"name"`
	Status    string          `json:"status"`
	Strategy  string          `json:"strategy"`
	Notes     string          `json:"notes"`
	Params    json.RawMessage `json:"params"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

type IndustryTask struct {
	ID            int64           `json:"id"`
	UserID        string          `json:"user_id"`
	ProjectID     int64           `json:"project_id"`
	ParentTaskID  int64           `json:"parent_task_id"`
	Name          string          `json:"name"`
	Activity      string          `json:"activity"`
	ProductTypeID int32           `json:"product_type_id"`
	TargetRuns    int32           `json:"target_runs"`
	PlannedStart  string          `json:"planned_start"`
	PlannedEnd    string          `json:"planned_end"`
	Priority      int             `json:"priority"`
	Status        string          `json:"status"`
	Constraints   json.RawMessage `json:"constraints"`
	CreatedAt     string          `json:"created_at"`
	UpdatedAt     string          `json:"updated_at"`
}

type IndustryJob struct {
	ID              int64   `json:"id"`
	UserID          string  `json:"user_id"`
	ProjectID       int64   `json:"project_id"`
	TaskID          int64   `json:"task_id"`
	CharacterID     int64   `json:"character_id"`
	FacilityID      int64   `json:"facility_id"`
	Activity        string  `json:"activity"`
	Runs            int32   `json:"runs"`
	DurationSeconds int64   `json:"duration_seconds"`
	CostISK         float64 `json:"cost_isk"`
	Status          string  `json:"status"`
	StartedAt       string  `json:"started_at"`
	FinishedAt      string  `json:"finished_at"`
	ExternalJobID   int64   `json:"external_job_id"`
	Notes           string  `json:"notes"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

type IndustryMaterialPlan struct {
	ID           int64   `json:"id"`
	UserID       string  `json:"user_id"`
	ProjectID    int64   `json:"project_id"`
	TaskID       int64   `json:"task_id"`
	TypeID       int32   `json:"type_id"`
	TypeName     string  `json:"type_name"`
	RequiredQty  int64   `json:"required_qty"`
	AvailableQty int64   `json:"available_qty"`
	BuyQty       int64   `json:"buy_qty"`
	BuildQty     int64   `json:"build_qty"`
	UnitCostISK  float64 `json:"unit_cost_isk"`
	Source       string  `json:"source"`
	UpdatedAt    string  `json:"updated_at"`
}

type IndustryBlueprintPool struct {
	ID              int64  `json:"id"`
	UserID          string `json:"user_id"`
	ProjectID       int64  `json:"project_id"`
	BlueprintTypeID int32  `json:"blueprint_type_id"`
	BlueprintName   string `json:"blueprint_name"`
	LocationID      int64  `json:"location_id"`
	Quantity        int64  `json:"quantity"`
	ME              int32  `json:"me"`
	TE              int32  `json:"te"`
	IsBPO           bool   `json:"is_bpo"`
	AvailableRuns   int64  `json:"available_runs"`
	UpdatedAt       string `json:"updated_at"`
}

type IndustryProjectCreateInput struct {
	Name     string      `json:"name"`
	Status   string      `json:"status"`
	Strategy string      `json:"strategy"`
	Notes    string      `json:"notes"`
	Params   interface{} `json:"params"`
}

type IndustryTaskPlanInput struct {
	TaskID        int64           `json:"task_id,omitempty"`
	ParentTaskID  int64           `json:"parent_task_id"`
	Name          string          `json:"name"`
	Activity      string          `json:"activity"`
	ProductTypeID int32           `json:"product_type_id"`
	TargetRuns    int32           `json:"target_runs"`
	PlannedStart  string          `json:"planned_start"`
	PlannedEnd    string          `json:"planned_end"`
	Priority      int             `json:"priority"`
	Status        string          `json:"status"`
	Constraints   json.RawMessage `json:"constraints"`
}

type IndustryJobPlanInput struct {
	TaskID          int64   `json:"task_id"`
	CharacterID     int64   `json:"character_id"`
	FacilityID      int64   `json:"facility_id"`
	Activity        string  `json:"activity"`
	Runs            int32   `json:"runs"`
	DurationSeconds int64   `json:"duration_seconds"`
	CostISK         float64 `json:"cost_isk"`
	Status          string  `json:"status"`
	StartedAt       string  `json:"started_at"`
	FinishedAt      string  `json:"finished_at"`
	ExternalJobID   int64   `json:"external_job_id"`
	Notes           string  `json:"notes"`
}

type IndustryMaterialPlanInput struct {
	TaskID       int64   `json:"task_id"`
	TypeID       int32   `json:"type_id"`
	TypeName     string  `json:"type_name"`
	RequiredQty  int64   `json:"required_qty"`
	AvailableQty int64   `json:"available_qty"`
	BuyQty       int64   `json:"buy_qty"`
	BuildQty     int64   `json:"build_qty"`
	UnitCostISK  float64 `json:"unit_cost_isk"`
	Source       string  `json:"source"`
}

type IndustryBlueprintPoolInput struct {
	BlueprintTypeID int32  `json:"blueprint_type_id"`
	BlueprintName   string `json:"blueprint_name"`
	LocationID      int64  `json:"location_id"`
	Quantity        int64  `json:"quantity"`
	ME              int32  `json:"me"`
	TE              int32  `json:"te"`
	IsBPO           bool   `json:"is_bpo"`
	AvailableRuns   int64  `json:"available_runs"`
}

type IndustryPlanSchedulerInput struct {
	Enabled               bool   `json:"enabled"`
	SlotCount             int    `json:"slot_count"`
	MaxJobRuns            int32  `json:"max_job_runs"`
	MaxJobDurationSeconds int64  `json:"max_job_duration_seconds"`
	WindowDays            int    `json:"window_days"`
	QueueStatus           string `json:"queue_status"`
}

type IndustryPlanPatch struct {
	Replace bool `json:"replace"`
	// ReplaceBlueprintPool clears blueprint rows for the project before upsert without touching tasks/jobs/materials.
	ReplaceBlueprintPool bool                         `json:"replace_blueprints"`
	ProjectStatus        string                       `json:"project_status"`
	Tasks                []IndustryTaskPlanInput      `json:"tasks"`
	Jobs                 []IndustryJobPlanInput       `json:"jobs"`
	Materials            []IndustryMaterialPlanInput  `json:"materials"`
	Blueprints           []IndustryBlueprintPoolInput `json:"blueprints"`
	Scheduler            IndustryPlanSchedulerInput   `json:"scheduler"`
	StrictBPBypass       bool                         `json:"strict_bp_bypass,omitempty"`
}

type IndustryPlanSummary struct {
	ProjectID        int64    `json:"project_id"`
	ProjectStatus    string   `json:"project_status"`
	Replaced         bool     `json:"replaced"`
	TasksInserted    int      `json:"tasks_inserted"`
	JobsInserted     int      `json:"jobs_inserted"`
	MaterialsUpsert  int      `json:"materials_upserted"`
	BlueprintsUpsert int      `json:"blueprints_upserted"`
	SchedulerApplied bool     `json:"scheduler_applied"`
	JobsSplitFrom    int      `json:"jobs_split_from"`
	JobsPlannedTotal int      `json:"jobs_planned_total"`
	Warnings         []string `json:"warnings,omitempty"`
	UpdatedAt        string   `json:"updated_at"`
}

type IndustryTaskPreview struct {
	InputIndex   int64  `json:"input_index"`
	TaskID       int64  `json:"task_id"`
	ParentTaskID int64  `json:"parent_task_id"`
	Name         string `json:"name"`
	Activity     string `json:"activity"`
	TargetRuns   int32  `json:"target_runs"`
	PlannedStart string `json:"planned_start"`
	PlannedEnd   string `json:"planned_end"`
	Priority     int    `json:"priority"`
}

type IndustryPlanPreview struct {
	ProjectID int64                  `json:"project_id"`
	Replace   bool                   `json:"replace"`
	Summary   IndustryPlanSummary    `json:"summary"`
	Tasks     []IndustryTaskPreview  `json:"tasks"`
	Jobs      []IndustryJobPlanInput `json:"jobs"`
	Warnings  []string               `json:"warnings"`
}

type IndustryLedgerEntry struct {
	JobID            int64   `json:"job_id"`
	ProjectID        int64   `json:"project_id"`
	ProjectName      string  `json:"project_name"`
	TaskID           int64   `json:"task_id"`
	TaskName         string  `json:"task_name"`
	CharacterID      int64   `json:"character_id"`
	FacilityID       int64   `json:"facility_id"`
	Activity         string  `json:"activity"`
	Runs             int32   `json:"runs"`
	DurationSeconds  int64   `json:"duration_seconds"`
	CostISK          float64 `json:"cost_isk"`
	Status           string  `json:"status"`
	StartedAt        string  `json:"started_at"`
	FinishedAt       string  `json:"finished_at"`
	ExternalJobID    int64   `json:"external_job_id"`
	Notes            string  `json:"notes"`
	LastUpdatedAtUTC string  `json:"updated_at"`
}

type IndustryLedger struct {
	ProjectID    int64                 `json:"project_id"`
	StatusFilter string                `json:"status_filter"`
	Limit        int                   `json:"limit"`
	Total        int64                 `json:"total"`
	Planned      int64                 `json:"planned"`
	Active       int64                 `json:"active"`
	Completed    int64                 `json:"completed"`
	Failed       int64                 `json:"failed"`
	Cancelled    int64                 `json:"cancelled"`
	TotalCostISK float64               `json:"total_cost_isk"`
	Entries      []IndustryLedgerEntry `json:"entries"`
}

type IndustryMaterialDiff struct {
	TypeID       int32  `json:"type_id"`
	TypeName     string `json:"type_name"`
	RequiredQty  int64  `json:"required_qty"`
	AvailableQty int64  `json:"available_qty"`
	BuyQty       int64  `json:"buy_qty"`
	BuildQty     int64  `json:"build_qty"`
	MissingQty   int64  `json:"missing_qty"`
}

type IndustryProjectSnapshot struct {
	Project      IndustryProject         `json:"project"`
	Tasks        []IndustryTask          `json:"tasks"`
	Jobs         []IndustryJob           `json:"jobs"`
	Materials    []IndustryMaterialPlan  `json:"materials"`
	Blueprints   []IndustryBlueprintPool `json:"blueprints"`
	MaterialDiff []IndustryMaterialDiff  `json:"material_diff"`
}

type IndustryLedgerOptions struct {
	ProjectID int64
	Status    string
	Limit     int
}

func normalizeIndustryProjectStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case IndustryProjectStatusDraft:
		return IndustryProjectStatusDraft
	case IndustryProjectStatusPlanned:
		return IndustryProjectStatusPlanned
	case IndustryProjectStatusActive:
		return IndustryProjectStatusActive
	case IndustryProjectStatusCompleted:
		return IndustryProjectStatusCompleted
	case IndustryProjectStatusArchived:
		return IndustryProjectStatusArchived
	default:
		return ""
	}
}

func defaultIndustryProjectStatus(v string) string {
	if normalized := normalizeIndustryProjectStatus(v); normalized != "" {
		return normalized
	}
	return IndustryProjectStatusDraft
}

func normalizeIndustryTaskStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case IndustryTaskStatusPlanned:
		return IndustryTaskStatusPlanned
	case IndustryTaskStatusReady:
		return IndustryTaskStatusReady
	case IndustryTaskStatusActive:
		return IndustryTaskStatusActive
	case IndustryTaskStatusPaused:
		return IndustryTaskStatusPaused
	case IndustryTaskStatusCompleted:
		return IndustryTaskStatusCompleted
	case IndustryTaskStatusBlocked:
		return IndustryTaskStatusBlocked
	case IndustryTaskStatusCancelled:
		return IndustryTaskStatusCancelled
	default:
		return ""
	}
}

func defaultIndustryTaskStatus(v string) string {
	if normalized := normalizeIndustryTaskStatus(v); normalized != "" {
		return normalized
	}
	return IndustryTaskStatusPlanned
}

func normalizeIndustryJobStatus(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case IndustryJobStatusPlanned:
		return IndustryJobStatusPlanned
	case IndustryJobStatusQueued:
		return IndustryJobStatusQueued
	case IndustryJobStatusActive:
		return IndustryJobStatusActive
	case IndustryJobStatusPaused:
		return IndustryJobStatusPaused
	case IndustryJobStatusCompleted:
		return IndustryJobStatusCompleted
	case IndustryJobStatusFailed:
		return IndustryJobStatusFailed
	case IndustryJobStatusCancelled:
		return IndustryJobStatusCancelled
	default:
		return ""
	}
}

func defaultIndustryJobStatus(v string) string {
	if normalized := normalizeIndustryJobStatus(v); normalized != "" {
		return normalized
	}
	return IndustryJobStatusPlanned
}

func normalizeIndustryStrategy(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "conservative":
		return "conservative"
	case "balanced":
		return "balanced"
	case "aggressive":
		return "aggressive"
	default:
		return "balanced"
	}
}

func normalizeIndustryActivity(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return "manufacturing"
	}
	return v
}

func normalizeIndustryMaterialSource(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "market", "stock", "build", "reprocess", "contract":
		return v
	default:
		return "market"
	}
}

func normalizeIndustryMaterialRebalanceStrategy(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "buy":
		return "buy"
	case "build":
		return "build"
	default:
		return "preserve"
	}
}

func normalizeIndustryMaterialWarehouseScope(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "global":
		return "global"
	case "strict_location":
		return "strict_location"
	default:
		return "location_first"
	}
}

func normalizeJSONRaw(raw json.RawMessage, fallback string) string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func marshalJSONOrDefault(v interface{}, fallback string) string {
	if v == nil {
		return fallback
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fallback
	}
	trimmed := strings.TrimSpace(string(b))
	if trimmed == "" || trimmed == "null" {
		return fallback
	}
	return trimmed
}

type industryMaterialStockPool struct {
	TotalQty   int64
	UnknownQty int64
	ByLocation map[int64]int64
}

func buildIndustryMaterialStockPools(
	stockByType map[int32]int64,
	stockByTypeLocation map[int32]map[int64]int64,
) map[int32]*industryMaterialStockPool {
	out := make(map[int32]*industryMaterialStockPool, len(stockByType)+len(stockByTypeLocation))

	for typeID, qty := range stockByType {
		if qty <= 0 {
			continue
		}
		out[typeID] = &industryMaterialStockPool{
			TotalQty:   qty,
			UnknownQty: qty,
			ByLocation: map[int64]int64{},
		}
	}

	for typeID, byLocation := range stockByTypeLocation {
		if len(byLocation) == 0 {
			continue
		}

		validLocationSum := int64(0)
		validLocations := make(map[int64]int64, len(byLocation))
		for locationID, qty := range byLocation {
			if locationID <= 0 || qty <= 0 {
				continue
			}
			validLocations[locationID] = qty
			validLocationSum += qty
		}
		if len(validLocations) == 0 {
			continue
		}

		pool := out[typeID]
		if pool == nil {
			pool = &industryMaterialStockPool{
				TotalQty:   validLocationSum,
				UnknownQty: validLocationSum,
				ByLocation: map[int64]int64{},
			}
			out[typeID] = pool
		}
		if pool.ByLocation == nil {
			pool.ByLocation = map[int64]int64{}
		}
		if pool.TotalQty <= 0 {
			pool.TotalQty = validLocationSum
			pool.UnknownQty = validLocationSum
		}

		locationIDs := make([]int64, 0, len(validLocations))
		for locationID := range validLocations {
			locationIDs = append(locationIDs, locationID)
		}
		sort.Slice(locationIDs, func(i, j int) bool { return locationIDs[i] < locationIDs[j] })

		remainingCap := pool.TotalQty
		for _, locationID := range locationIDs {
			if remainingCap <= 0 {
				break
			}
			qty := validLocations[locationID]
			if qty <= 0 {
				continue
			}
			if qty > remainingCap {
				qty = remainingCap
			}
			pool.ByLocation[locationID] = qty
			remainingCap -= qty
		}
		pool.UnknownQty = remainingCap
	}

	return out
}

func industryMaterialAllocateFromLocation(pool *industryMaterialStockPool, locationID int64, want int64) int64 {
	if pool == nil || locationID <= 0 || want <= 0 || pool.TotalQty <= 0 {
		return 0
	}
	if pool.ByLocation == nil {
		return 0
	}
	locationQty := pool.ByLocation[locationID]
	if locationQty <= 0 {
		return 0
	}

	alloc := want
	if locationQty < alloc {
		alloc = locationQty
	}
	if pool.TotalQty < alloc {
		alloc = pool.TotalQty
	}
	if alloc <= 0 {
		return 0
	}

	locationQty -= alloc
	if locationQty > 0 {
		pool.ByLocation[locationID] = locationQty
	} else {
		delete(pool.ByLocation, locationID)
	}
	pool.TotalQty -= alloc
	return alloc
}

func industryMaterialAllocateGlobal(pool *industryMaterialStockPool, want int64) int64 {
	if pool == nil || want <= 0 || pool.TotalQty <= 0 {
		return 0
	}

	need := want
	if pool.TotalQty < need {
		need = pool.TotalQty
	}
	if need <= 0 {
		return 0
	}

	allocated := int64(0)
	if pool.UnknownQty > 0 {
		take := need
		if pool.UnknownQty < take {
			take = pool.UnknownQty
		}
		pool.UnknownQty -= take
		pool.TotalQty -= take
		need -= take
		allocated += take
	}

	if need <= 0 {
		return allocated
	}

	locationIDs := make([]int64, 0, len(pool.ByLocation))
	for locationID := range pool.ByLocation {
		locationIDs = append(locationIDs, locationID)
	}
	sort.Slice(locationIDs, func(i, j int) bool { return locationIDs[i] < locationIDs[j] })

	for _, locationID := range locationIDs {
		if need <= 0 {
			break
		}
		locationQty := pool.ByLocation[locationID]
		if locationQty <= 0 {
			continue
		}

		take := need
		if locationQty < take {
			take = locationQty
		}
		pool.TotalQty -= take
		need -= take
		allocated += take
		locationQty -= take
		if locationQty > 0 {
			pool.ByLocation[locationID] = locationQty
		} else {
			delete(pool.ByLocation, locationID)
		}
	}

	return allocated
}

type industrySchedulerConfig struct {
	Enabled               bool
	SlotCount             int
	MaxJobRuns            int32
	MaxJobDurationSeconds int64
	WindowDays            int
	QueueStatus           string
}

type industryPlanTaskRecord struct {
	InputIndex      int64
	SourceTaskID    int64
	ID              int64
	ParentTaskID    int64
	Name            string
	Activity        string
	TargetRuns      int32
	PlannedStart    string
	PlannedEnd      string
	Priority        int
	ConstraintsJSON string
}

func defaultIndustrySchedulerProfile(strategy string) industrySchedulerConfig {
	switch normalizeIndustryStrategy(strategy) {
	case "conservative":
		return industrySchedulerConfig{
			Enabled:               false,
			SlotCount:             1,
			MaxJobRuns:            50,
			MaxJobDurationSeconds: 12 * 3600,
			WindowDays:            30,
			QueueStatus:           IndustryJobStatusPlanned,
		}
	case "aggressive":
		return industrySchedulerConfig{
			Enabled:               false,
			SlotCount:             4,
			MaxJobRuns:            400,
			MaxJobDurationSeconds: 72 * 3600,
			WindowDays:            30,
			QueueStatus:           IndustryJobStatusQueued,
		}
	default:
		return industrySchedulerConfig{
			Enabled:               false,
			SlotCount:             2,
			MaxJobRuns:            200,
			MaxJobDurationSeconds: 24 * 3600,
			WindowDays:            30,
			QueueStatus:           IndustryJobStatusQueued,
		}
	}
}

func normalizeIndustrySchedulerInput(in IndustryPlanSchedulerInput, strategy string) industrySchedulerConfig {
	profile := defaultIndustrySchedulerProfile(strategy)
	cfg := industrySchedulerConfig{
		Enabled:               in.Enabled,
		SlotCount:             profile.SlotCount,
		MaxJobRuns:            profile.MaxJobRuns,
		MaxJobDurationSeconds: profile.MaxJobDurationSeconds,
		WindowDays:            profile.WindowDays,
		QueueStatus:           profile.QueueStatus,
	}
	if in.SlotCount > 0 {
		cfg.SlotCount = in.SlotCount
	}
	if in.MaxJobRuns > 0 {
		cfg.MaxJobRuns = in.MaxJobRuns
	}
	if in.MaxJobDurationSeconds > 0 {
		cfg.MaxJobDurationSeconds = in.MaxJobDurationSeconds
	}
	if normalizedQueue := normalizeIndustryJobStatus(in.QueueStatus); normalizedQueue != "" {
		cfg.QueueStatus = normalizedQueue
	}
	if cfg.SlotCount <= 0 {
		cfg.SlotCount = profile.SlotCount
	}
	if cfg.MaxJobRuns <= 0 {
		cfg.MaxJobRuns = profile.MaxJobRuns
	}
	if cfg.MaxJobDurationSeconds <= 0 {
		cfg.MaxJobDurationSeconds = profile.MaxJobDurationSeconds
	}
	windowDays := in.WindowDays
	if windowDays <= 0 {
		windowDays = profile.WindowDays
	}
	if windowDays > 30 {
		windowDays = 30
	}
	cfg.WindowDays = windowDays
	windowCapSeconds := int64(windowDays) * 24 * 3600
	if cfg.MaxJobDurationSeconds <= 0 {
		cfg.MaxJobDurationSeconds = windowCapSeconds
	} else if cfg.MaxJobDurationSeconds > windowCapSeconds {
		cfg.MaxJobDurationSeconds = windowCapSeconds
	}
	if cfg.QueueStatus == "" {
		cfg.QueueStatus = profile.QueueStatus
	}
	return cfg
}

func parseRFC3339UTC(v string) (time.Time, bool) {
	trimmed := strings.TrimSpace(v)
	if trimmed == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return time.Time{}, false
	}
	return t.UTC(), true
}

func normalizeOptionalRFC3339(v string, fieldName string) (string, error) {
	trimmed := strings.TrimSpace(v)
	if trimmed == "" {
		return "", nil
	}
	parsed, ok := parseRFC3339UTC(trimmed)
	if !ok {
		return "", fmt.Errorf("%s must be RFC3339", fieldName)
	}
	return parsed.Format(time.RFC3339), nil
}

const industryTaskRefAmbiguityWarning = "ambiguous task references kept as existing IDs; use negative row refs (-1, -2, ...) for draft task links"

func buildIndustrySourceTaskIDMap(taskRecords []industryPlanTaskRecord) map[int64]int64 {
	if len(taskRecords) == 0 {
		return map[int64]int64{}
	}
	out := make(map[int64]int64, len(taskRecords))
	for _, rec := range taskRecords {
		if rec.SourceTaskID <= 0 || rec.ID <= 0 {
			continue
		}
		if _, exists := out[rec.SourceTaskID]; exists {
			continue
		}
		out[rec.SourceTaskID] = rec.ID
	}
	return out
}

func remapIndustryTaskReference(
	ref int64,
	inputTaskIndexToID map[int64]int64,
	sourceTaskIDToID map[int64]int64,
	patchTaskCount int,
	replace bool,
	existingTaskIDs map[int64]struct{},
) (int64, bool) {
	if ref == 0 {
		return ref, false
	}

	inputRef := ref
	forceInputRef := false
	if inputRef < 0 {
		inputRef = -inputRef
		forceInputRef = true
	}

	if forceInputRef {
		if inputRef <= 0 {
			return ref, false
		}
		if mapped, ok := inputTaskIndexToID[inputRef]; ok {
			return mapped, false
		}
		return ref, false
	}

	if mapped, ok := sourceTaskIDToID[ref]; ok {
		if !replace {
			if _, exists := existingTaskIDs[ref]; exists {
				return ref, true
			}
		}
		return mapped, false
	}

	if !replace {
		if _, exists := existingTaskIDs[ref]; exists {
			return ref, false
		}
	}

	maxInputRef := int64(patchTaskCount)
	if maxInputRef > 0 && inputRef > 0 && inputRef <= maxInputRef {
		if mapped, ok := inputTaskIndexToID[inputRef]; ok {
			// Positive row refs are legacy/ambiguous in append mode.
			return mapped, !replace
		}
	}

	return ref, false
}

func remapIndustryJobTaskIDs(
	jobs []IndustryJobPlanInput,
	inputTaskIndexToID map[int64]int64,
	sourceTaskIDToID map[int64]int64,
	patchTaskCount int,
	replace bool,
	existingTaskIDs map[int64]struct{},
) int {
	if len(jobs) == 0 {
		return 0
	}
	ambiguousRefs := 0
	for i := range jobs {
		if jobs[i].TaskID == 0 {
			continue
		}
		remappedID, ambiguous := remapIndustryTaskReference(
			jobs[i].TaskID,
			inputTaskIndexToID,
			sourceTaskIDToID,
			patchTaskCount,
			replace,
			existingTaskIDs,
		)
		jobs[i].TaskID = remappedID
		if ambiguous {
			ambiguousRefs++
		}
	}
	return ambiguousRefs
}

func nullablePositiveInt64(v int64) interface{} {
	if v <= 0 {
		return nil
	}
	return v
}

func normalizeIndustryBlueprintAvailableRuns(isBPO bool, availableRuns int64) int64 {
	if isBPO {
		return 0
	}
	if availableRuns < 0 {
		return 0
	}
	return availableRuns
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func jobTerminalStatus(status string) bool {
	switch status {
	case IndustryJobStatusActive, IndustryJobStatusCompleted, IndustryJobStatusFailed, IndustryJobStatusCancelled:
		return true
	default:
		return false
	}
}

func extractConstraintFloat(constraints map[string]interface{}, keys ...string) (float64, bool) {
	for _, key := range keys {
		raw, ok := constraints[key]
		if !ok {
			continue
		}
		switch v := raw.(type) {
		case float64:
			return v, true
		case int64:
			return float64(v), true
		case int32:
			return float64(v), true
		case int:
			return float64(v), true
		}
	}
	return 0, false
}

func extractConstraintInt64(constraints map[string]interface{}, keys ...string) (int64, bool) {
	for _, key := range keys {
		raw, ok := constraints[key]
		if !ok {
			continue
		}
		switch v := raw.(type) {
		case float64:
			return int64(math.Round(v)), true
		case int64:
			return v, true
		case int32:
			return int64(v), true
		case int:
			return int64(v), true
		case json.Number:
			if parsed, err := v.Int64(); err == nil {
				return parsed, true
			}
			if parsed, err := v.Float64(); err == nil {
				return int64(math.Round(parsed)), true
			}
		case string:
			if parsed, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil {
				return parsed, true
			}
		}
	}
	return 0, false
}

func listIndustryTaskPreferredLocationForProjectTx(
	tx *sql.Tx,
	userID string,
	projectID int64,
) (map[int64]int64, error) {
	rows, err := tx.Query(`
		SELECT id, constraints_json
		  FROM industry_tasks
		 WHERE user_id = ? AND project_id = ?
	`, userID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[int64]int64, 128)
	for rows.Next() {
		var (
			taskID          int64
			constraintsJSON string
		)
		if err := rows.Scan(&taskID, &constraintsJSON); err != nil {
			return nil, err
		}
		if taskID <= 0 {
			continue
		}

		constraintsJSON = strings.TrimSpace(constraintsJSON)
		if constraintsJSON == "" {
			continue
		}

		constraints := map[string]interface{}{}
		if err := json.Unmarshal([]byte(constraintsJSON), &constraints); err != nil {
			continue
		}
		locationID, ok := extractConstraintInt64(
			constraints,
			"warehouse_location_id",
			"blueprint_location_id",
			"bp_location_id",
			"station_id",
			"facility_id",
			"location_id",
		)
		if !ok || locationID <= 0 {
			continue
		}
		out[taskID] = locationID
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

type industryTaskBlueprintConstraint struct {
	TaskID              int64
	TaskName            string
	BlueprintTypeID     int32
	BlueprintLocationID int64
}

func buildIndustryTaskBlueprintConstraints(tasks []industryPlanTaskRecord) map[int64]industryTaskBlueprintConstraint {
	out := make(map[int64]industryTaskBlueprintConstraint, len(tasks))
	for _, task := range tasks {
		constraintsJSON := strings.TrimSpace(task.ConstraintsJSON)
		if constraintsJSON == "" {
			continue
		}
		constraints := map[string]interface{}{}
		if err := json.Unmarshal([]byte(constraintsJSON), &constraints); err != nil {
			continue
		}
		blueprintTypeID, ok := extractConstraintInt64(
			constraints,
			"blueprint_type_id",
			"bp_type_id",
		)
		if !ok || blueprintTypeID <= 0 {
			continue
		}
		blueprintLocationID, _ := extractConstraintInt64(
			constraints,
			"blueprint_location_id",
			"bp_location_id",
		)
		if blueprintLocationID <= 0 {
			if stationID, hasStation := extractConstraintInt64(
				constraints,
				"station_id",
				"location_id",
			); hasStation && stationID > 0 {
				blueprintLocationID = stationID
			}
		}
		out[task.ID] = industryTaskBlueprintConstraint{
			TaskID:              task.ID,
			TaskName:            strings.TrimSpace(task.Name),
			BlueprintTypeID:     int32(blueprintTypeID),
			BlueprintLocationID: blueprintLocationID,
		}
	}
	return out
}

func (d *DB) listIndustryBlueprintPoolForProject(
	userID string,
	projectID int64,
) ([]IndustryBlueprintPool, error) {
	rows, err := d.sql.Query(`
		SELECT id, user_id, project_id, blueprint_type_id, blueprint_name, location_id,
		       quantity, me, te, is_bpo, available_runs, updated_at
		  FROM industry_blueprint_pool
		 WHERE user_id = ? AND project_id = ?
	`, userID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIndustryBlueprintPoolRows(rows)
}

func listIndustryBlueprintPoolForProjectTx(
	tx *sql.Tx,
	userID string,
	projectID int64,
) ([]IndustryBlueprintPool, error) {
	rows, err := tx.Query(`
		SELECT id, user_id, project_id, blueprint_type_id, blueprint_name, location_id,
		       quantity, me, te, is_bpo, available_runs, updated_at
		  FROM industry_blueprint_pool
		 WHERE user_id = ? AND project_id = ?
	`, userID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIndustryBlueprintPoolRows(rows)
}

func scanIndustryBlueprintPoolRows(rows *sql.Rows) ([]IndustryBlueprintPool, error) {
	out := make([]IndustryBlueprintPool, 0, 32)
	for rows.Next() {
		var (
			row   IndustryBlueprintPool
			isBPO int
		)
		if err := rows.Scan(
			&row.ID,
			&row.UserID,
			&row.ProjectID,
			&row.BlueprintTypeID,
			&row.BlueprintName,
			&row.LocationID,
			&row.Quantity,
			&row.ME,
			&row.TE,
			&isBPO,
			&row.AvailableRuns,
			&row.UpdatedAt,
		); err != nil {
			return nil, err
		}
		row.IsBPO = isBPO > 0
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (d *DB) listIndustryTaskRecordsForProject(
	userID string,
	projectID int64,
) ([]industryPlanTaskRecord, error) {
	rows, err := d.sql.Query(`
		SELECT id, name, activity, target_runs, planned_start, planned_end, priority, constraints_json
		  FROM industry_tasks
		 WHERE user_id = ? AND project_id = ?
	`, userID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIndustryTaskRecords(rows)
}

func listIndustryTaskRecordsForProjectTx(
	tx *sql.Tx,
	userID string,
	projectID int64,
) ([]industryPlanTaskRecord, error) {
	rows, err := tx.Query(`
		SELECT id, name, activity, target_runs, planned_start, planned_end, priority, constraints_json
		  FROM industry_tasks
		 WHERE user_id = ? AND project_id = ?
	`, userID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIndustryTaskRecords(rows)
}

func scanIndustryTaskRecords(rows *sql.Rows) ([]industryPlanTaskRecord, error) {
	out := make([]industryPlanTaskRecord, 0, 64)
	for rows.Next() {
		var row industryPlanTaskRecord
		if err := rows.Scan(
			&row.ID,
			&row.Name,
			&row.Activity,
			&row.TargetRuns,
			&row.PlannedStart,
			&row.PlannedEnd,
			&row.Priority,
			&row.ConstraintsJSON,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func mergeIndustryBlueprintPool(
	base []IndustryBlueprintPool,
	patch []IndustryBlueprintPoolInput,
) []IndustryBlueprintPool {
	type bpKey struct {
		TypeID     int32
		LocationID int64
		IsBPO      bool
	}
	byKey := make(map[bpKey]IndustryBlueprintPool, len(base)+len(patch))
	for _, row := range base {
		key := bpKey{
			TypeID:     row.BlueprintTypeID,
			LocationID: row.LocationID,
			IsBPO:      row.IsBPO,
		}
		byKey[key] = row
	}

	for _, in := range patch {
		if in.BlueprintTypeID <= 0 {
			continue
		}
		key := bpKey{
			TypeID:     in.BlueprintTypeID,
			LocationID: in.LocationID,
			IsBPO:      in.IsBPO,
		}
		row := byKey[key]
		row.BlueprintTypeID = in.BlueprintTypeID
		row.BlueprintName = strings.TrimSpace(in.BlueprintName)
		row.LocationID = in.LocationID
		row.Quantity = in.Quantity
		row.ME = in.ME
		row.TE = in.TE
		row.IsBPO = in.IsBPO
		row.AvailableRuns = normalizeIndustryBlueprintAvailableRuns(in.IsBPO, in.AvailableRuns)
		byKey[key] = row
	}

	out := make([]IndustryBlueprintPool, 0, len(byKey))
	for _, row := range byKey {
		out = append(out, row)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].BlueprintTypeID != out[j].BlueprintTypeID {
			return out[i].BlueprintTypeID < out[j].BlueprintTypeID
		}
		if out[i].LocationID != out[j].LocationID {
			return out[i].LocationID < out[j].LocationID
		}
		if out[i].IsBPO == out[j].IsBPO {
			return out[i].ID < out[j].ID
		}
		return out[i].IsBPO && !out[j].IsBPO
	})
	return out
}

type industryBlueprintCapacity struct {
	CapRuns         int64
	HasTypeInPool   bool
	HasMatchingPool bool
	Unlimited       bool
}

func industryBlueprintCapacityForTask(
	pool []IndustryBlueprintPool,
	blueprintTypeID int32,
	locationID int64,
) industryBlueprintCapacity {
	out := industryBlueprintCapacity{}
	for _, row := range pool {
		if row.BlueprintTypeID != blueprintTypeID {
			continue
		}
		out.HasTypeInPool = true
		if locationID > 0 && row.LocationID > 0 && row.LocationID != locationID {
			continue
		}
		out.HasMatchingPool = true

		// BPO capacity is intentionally unbounded; AvailableRuns is ignored for BPO rows.
		if row.IsBPO {
			out.Unlimited = true
			out.CapRuns = 0
			return out
		}
		if row.AvailableRuns > 0 {
			out.CapRuns += row.AvailableRuns
		}
	}
	return out
}

func applyBlueprintRunCapsToJobs(
	jobs []IndustryJobPlanInput,
	tasks []industryPlanTaskRecord,
	blueprintPool []IndustryBlueprintPool,
	strictGate bool,
) ([]IndustryJobPlanInput, []string) {
	if len(jobs) == 0 || len(tasks) == 0 {
		return jobs, nil
	}

	taskConstraints := buildIndustryTaskBlueprintConstraints(tasks)
	if len(taskConstraints) == 0 {
		return jobs, nil
	}

	demandByTask := make(map[int64]int64, len(taskConstraints))
	for _, job := range jobs {
		if _, ok := taskConstraints[job.TaskID]; !ok {
			continue
		}
		if job.Runs <= 0 {
			continue
		}
		demandByTask[job.TaskID] += int64(job.Runs)
	}

	remainingByTask := make(map[int64]int64, len(demandByTask))
	warnings := make([]string, 0, 8)
	for taskID, demandRuns := range demandByTask {
		constraint := taskConstraints[taskID]
		capacity := industryBlueprintCapacityForTask(
			blueprintPool,
			constraint.BlueprintTypeID,
			constraint.BlueprintLocationID,
		)

		if !capacity.HasMatchingPool {
			if strictGate {
				remainingByTask[taskID] = 0
				taskName := constraint.TaskName
				if taskName == "" {
					taskName = fmt.Sprintf("task-%d", taskID)
				}
				if constraint.BlueprintLocationID > 0 && capacity.HasTypeInPool {
					warnings = append(warnings, fmt.Sprintf(
						"blueprint gate: %s (%d) missing blueprint at location %d (bp=%d, demand=%d)",
						taskName,
						taskID,
						constraint.BlueprintLocationID,
						constraint.BlueprintTypeID,
						demandRuns,
					))
				} else {
					warnings = append(warnings, fmt.Sprintf(
						"blueprint gate: %s (%d) missing owned blueprint (bp=%d, demand=%d)",
						taskName,
						taskID,
						constraint.BlueprintTypeID,
						demandRuns,
					))
				}
			}
			continue
		}

		if capacity.Unlimited || demandRuns <= capacity.CapRuns {
			continue
		}
		remainingByTask[taskID] = capacity.CapRuns
		taskName := constraint.TaskName
		if taskName == "" {
			taskName = fmt.Sprintf("task-%d", taskID)
		}
		warnings = append(warnings, fmt.Sprintf(
			"blueprint cap: %s (%d) limited to %d/%d runs (bp=%d, shortage=%d)",
			taskName,
			taskID,
			capacity.CapRuns,
			demandRuns,
			constraint.BlueprintTypeID,
			demandRuns-capacity.CapRuns,
		))
	}
	if len(remainingByTask) == 0 {
		return jobs, warnings
	}

	out := make([]IndustryJobPlanInput, 0, len(jobs))
	for _, job := range jobs {
		remaining, limited := remainingByTask[job.TaskID]
		if !limited {
			out = append(out, job)
			continue
		}
		jobRuns := int64(job.Runs)
		if jobRuns <= 0 {
			continue
		}
		if remaining <= 0 {
			continue
		}
		if jobRuns <= remaining {
			remainingByTask[job.TaskID] = remaining - jobRuns
			out = append(out, job)
			continue
		}

		newRuns := remaining
		ratio := float64(newRuns) / float64(jobRuns)
		if job.DurationSeconds > 0 {
			job.DurationSeconds = int64(math.Round(float64(job.DurationSeconds) * ratio))
		}
		if job.CostISK > 0 {
			job.CostISK = job.CostISK * ratio
		}
		job.Runs = int32(newRuns)
		if constraint, ok := taskConstraints[job.TaskID]; ok {
			notes := strings.TrimSpace(job.Notes)
			if notes != "" {
				notes += " | "
			}
			notes += fmt.Sprintf(
				"blueprint capped %d/%d runs bp=%d",
				newRuns,
				jobRuns,
				constraint.BlueprintTypeID,
			)
			job.Notes = notes
		}
		remainingByTask[job.TaskID] = 0
		out = append(out, job)
	}

	return out, warnings
}

func deriveIndustryJobsFromTaskRecords(tasks []industryPlanTaskRecord, queueStatus string) []IndustryJobPlanInput {
	out := make([]IndustryJobPlanInput, 0, len(tasks))
	for _, task := range tasks {
		runs := task.TargetRuns
		if runs <= 0 {
			continue
		}
		job := IndustryJobPlanInput{
			TaskID:        task.ID,
			Activity:      normalizeIndustryActivity(task.Activity),
			Runs:          runs,
			Status:        defaultIndustryJobStatus(queueStatus),
			StartedAt:     strings.TrimSpace(task.PlannedStart),
			FinishedAt:    strings.TrimSpace(task.PlannedEnd),
			ExternalJobID: 0,
			Notes:         strings.TrimSpace(task.Name),
		}
		constraints := map[string]interface{}{}
		if strings.TrimSpace(task.ConstraintsJSON) != "" {
			_ = json.Unmarshal([]byte(task.ConstraintsJSON), &constraints)
		}
		if perRunSeconds, ok := extractConstraintFloat(constraints, "duration_seconds_per_run"); ok && perRunSeconds > 0 {
			job.DurationSeconds = int64(math.Round(perRunSeconds * float64(runs)))
		} else if totalSeconds, ok := extractConstraintFloat(constraints, "duration_seconds"); ok && totalSeconds > 0 {
			job.DurationSeconds = int64(math.Round(totalSeconds))
		}
		if perRunCost, ok := extractConstraintFloat(constraints, "cost_isk_per_run"); ok && perRunCost > 0 {
			job.CostISK = perRunCost * float64(runs)
		} else if totalCost, ok := extractConstraintFloat(constraints, "cost_isk"); ok && totalCost > 0 {
			job.CostISK = totalCost
		}
		out = append(out, job)
	}
	return out
}

func splitAndScheduleIndustryJobs(
	jobs []IndustryJobPlanInput,
	cfg industrySchedulerConfig,
	now time.Time,
	taskParents map[int64]int64,
	taskPlannedEnd map[int64]time.Time,
) []IndustryJobPlanInput {
	if len(jobs) == 0 {
		return []IndustryJobPlanInput{}
	}
	type jobWorkItem struct {
		Index int
		Depth int
		Job   IndustryJobPlanInput
	}
	depthCache := map[int64]int{}
	var depth func(taskID int64, visiting map[int64]bool) int
	depth = func(taskID int64, visiting map[int64]bool) int {
		if taskID <= 0 {
			return 0
		}
		if d, ok := depthCache[taskID]; ok {
			return d
		}
		if visiting[taskID] {
			return 0
		}
		visiting[taskID] = true
		d := 0
		if parentID := taskParents[taskID]; parentID > 0 {
			d = 1 + depth(parentID, visiting)
		}
		depthCache[taskID] = d
		delete(visiting, taskID)
		return d
	}
	work := make([]jobWorkItem, 0, len(jobs))
	for i, job := range jobs {
		work = append(work, jobWorkItem{
			Index: i,
			Depth: depth(job.TaskID, map[int64]bool{}),
			Job:   job,
		})
	}
	sort.SliceStable(work, func(i, j int) bool {
		if work[i].Depth != work[j].Depth {
			return work[i].Depth < work[j].Depth
		}
		return work[i].Index < work[j].Index
	})

	slots := make([]time.Time, cfg.SlotCount)
	for i := range slots {
		slots[i] = now
	}
	windowDays := cfg.WindowDays
	if windowDays <= 0 {
		windowDays = 30
	}
	horizonEnd := now.Add(time.Duration(windowDays) * 24 * time.Hour)

	out := make([]IndustryJobPlanInput, 0, len(jobs))
	taskLatestFinish := map[int64]time.Time{}

	dependencyReadyAt := func(taskID int64) time.Time {
		var walk func(tid int64, seen map[int64]bool) time.Time
		walk = func(tid int64, seen map[int64]bool) time.Time {
			parentID := taskParents[tid]
			if parentID <= 0 {
				return time.Time{}
			}
			if seen[parentID] {
				return time.Time{}
			}
			seen[parentID] = true
			anchor := walk(parentID, seen)
			if plannedEnd, ok := taskPlannedEnd[parentID]; ok && plannedEnd.After(anchor) {
				anchor = plannedEnd
			}
			if finished, ok := taskLatestFinish[parentID]; ok && finished.After(anchor) {
				anchor = finished
			}
			return anchor
		}
		return walk(taskID, map[int64]bool{})
	}

	resolveStatus := func(baseStatus string) string {
		status := defaultIndustryJobStatus(baseStatus)
		if !jobTerminalStatus(status) {
			status = cfg.QueueStatus
		}
		return status
	}

	for _, wi := range work {
		jobIdx := wi.Index
		base := wi.Job
		baseRuns := base.Runs
		if baseRuns <= 0 {
			baseRuns = 1
		}
		baseDuration := base.DurationSeconds
		perRunDuration := int64(0)
		if baseDuration > 0 && baseRuns > 0 {
			perRunDuration = int64(math.Ceil(float64(baseDuration) / float64(baseRuns)))
		}
		perRunCost := 0.0
		if baseRuns > 0 && base.CostISK > 0 {
			perRunCost = base.CostISK / float64(baseRuns)
		}

		maxRuns := cfg.MaxJobRuns
		if maxRuns <= 0 {
			maxRuns = 1
		}
		if perRunDuration > 0 && cfg.MaxJobDurationSeconds > 0 {
			allowedByDuration := int32(cfg.MaxJobDurationSeconds / perRunDuration)
			if allowedByDuration < 1 {
				allowedByDuration = 1
			}
			if allowedByDuration < maxRuns {
				maxRuns = allowedByDuration
			}
		}
		if maxRuns < 1 {
			maxRuns = 1
		}

		baseAnchor := now
		if parsed, ok := parseRFC3339UTC(base.StartedAt); ok {
			baseAnchor = parsed
		}
		if depAnchor := dependencyReadyAt(base.TaskID); depAnchor.After(baseAnchor) {
			baseAnchor = depAnchor
		}

		remaining := baseRuns
		chunkNum := int32(0)
		totalChunks := int32(math.Ceil(float64(baseRuns) / float64(maxRuns)))
		if totalChunks < 1 {
			totalChunks = 1
		}

		appendDeferredChunk := func(runs int32) {
			if runs <= 0 {
				return
			}
			notes := strings.TrimSpace(base.Notes)
			if notes != "" {
				notes += " | "
			}
			notes += fmt.Sprintf("scheduler deferred %d runs beyond %d-day horizon source_job=%d", runs, windowDays, jobIdx+1)
			deferredDuration := int64(0)
			if perRunDuration > 0 {
				deferredDuration = perRunDuration * int64(runs)
			}
			out = append(out, IndustryJobPlanInput{
				TaskID:          base.TaskID,
				CharacterID:     base.CharacterID,
				FacilityID:      base.FacilityID,
				Activity:        normalizeIndustryActivity(base.Activity),
				Runs:            runs,
				DurationSeconds: deferredDuration,
				CostISK:         perRunCost * float64(runs),
				Status:          resolveStatus(base.Status),
				StartedAt:       "",
				FinishedAt:      "",
				ExternalJobID:   base.ExternalJobID,
				Notes:           notes,
			})
			if base.TaskID > 0 {
				if current, ok := taskLatestFinish[base.TaskID]; !ok || horizonEnd.After(current) {
					taskLatestFinish[base.TaskID] = horizonEnd
				}
			}
		}

		for remaining > 0 {
			chunkRuns := maxRuns
			if remaining < chunkRuns {
				chunkRuns = remaining
			}

			slotIdx := 0
			for i := 1; i < len(slots); i++ {
				if slots[i].Before(slots[slotIdx]) {
					slotIdx = i
				}
			}
			startAt := maxTime(slots[slotIdx], baseAnchor)
			if !startAt.Before(horizonEnd) {
				appendDeferredChunk(remaining)
				remaining = 0
				break
			}
			if perRunDuration > 0 {
				secondsLeft := int64(horizonEnd.Sub(startAt).Seconds())
				if secondsLeft < perRunDuration {
					appendDeferredChunk(remaining)
					remaining = 0
					break
				}
				allowedByHorizon := int32(secondsLeft / perRunDuration)
				if allowedByHorizon < 1 {
					appendDeferredChunk(remaining)
					remaining = 0
					break
				}
				if allowedByHorizon < chunkRuns {
					chunkRuns = allowedByHorizon
				}
			}
			chunkNum++
			remaining -= chunkRuns

			chunkDuration := int64(0)
			if perRunDuration > 0 {
				chunkDuration = perRunDuration * int64(chunkRuns)
			}
			finishAt := startAt
			if chunkDuration > 0 {
				finishAt = startAt.Add(time.Duration(chunkDuration) * time.Second)
				slots[slotIdx] = finishAt
			} else {
				slots[slotIdx] = startAt
			}

			status := resolveStatus(base.Status)
			notes := strings.TrimSpace(base.Notes)
			if notes != "" {
				notes += " | "
			}
			notes += fmt.Sprintf("scheduler chunk %d/%d slot=%d source_job=%d", chunkNum, totalChunks, slotIdx+1, jobIdx+1)

			scheduled := IndustryJobPlanInput{
				TaskID:          base.TaskID,
				CharacterID:     base.CharacterID,
				FacilityID:      base.FacilityID,
				Activity:        normalizeIndustryActivity(base.Activity),
				Runs:            chunkRuns,
				DurationSeconds: chunkDuration,
				CostISK:         perRunCost * float64(chunkRuns),
				Status:          status,
				StartedAt:       startAt.Format(time.RFC3339),
				FinishedAt:      "",
				ExternalJobID:   base.ExternalJobID,
				Notes:           notes,
			}
			if chunkDuration > 0 {
				scheduled.FinishedAt = finishAt.Format(time.RFC3339)
			} else if parsed, ok := parseRFC3339UTC(base.FinishedAt); ok {
				scheduled.FinishedAt = parsed.Format(time.RFC3339)
			}
			if base.TaskID > 0 {
				mark := startAt
				if chunkDuration > 0 {
					mark = finishAt
				}
				if current, ok := taskLatestFinish[base.TaskID]; !ok || mark.After(current) {
					taskLatestFinish[base.TaskID] = mark
				}
			}
			out = append(out, scheduled)
		}
	}
	return out
}

func loadIndustryTaskSchedulingMapsTx(
	tx *sql.Tx,
	userID string,
	projectID int64,
	taskParents map[int64]int64,
	taskPlannedEnd map[int64]time.Time,
) error {
	rows, err := tx.Query(`
		SELECT id, COALESCE(parent_task_id, 0), planned_end
		  FROM industry_tasks
		 WHERE user_id = ? AND project_id = ?
	`, userID, projectID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			taskID     int64
			parentID   int64
			plannedEnd string
		)
		if err := rows.Scan(&taskID, &parentID, &plannedEnd); err != nil {
			return err
		}
		taskParents[taskID] = parentID
		if t, ok := parseRFC3339UTC(plannedEnd); ok {
			taskPlannedEnd[taskID] = t
		}
	}
	return rows.Err()
}

func loadIndustryTaskIDSetTx(tx *sql.Tx, userID string, projectID int64) (map[int64]struct{}, error) {
	rows, err := tx.Query(`
		SELECT id
		  FROM industry_tasks
		 WHERE user_id = ? AND project_id = ?
	`, userID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[int64]struct{}{}
	for rows.Next() {
		var taskID int64
		if err := rows.Scan(&taskID); err != nil {
			return nil, err
		}
		out[taskID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (d *DB) loadIndustryTaskSchedulingMaps(
	userID string,
	projectID int64,
	taskParents map[int64]int64,
	taskPlannedEnd map[int64]time.Time,
) error {
	rows, err := d.sql.Query(`
		SELECT id, COALESCE(parent_task_id, 0), planned_end
		  FROM industry_tasks
		 WHERE user_id = ? AND project_id = ?
	`, userID, projectID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			taskID     int64
			parentID   int64
			plannedEnd string
		)
		if err := rows.Scan(&taskID, &parentID, &plannedEnd); err != nil {
			return err
		}
		taskParents[taskID] = parentID
		if t, ok := parseRFC3339UTC(plannedEnd); ok {
			taskPlannedEnd[taskID] = t
		}
	}
	return rows.Err()
}

func (d *DB) loadIndustryTaskIDSet(userID string, projectID int64) (map[int64]struct{}, error) {
	rows, err := d.sql.Query(`
		SELECT id
		  FROM industry_tasks
		 WHERE user_id = ? AND project_id = ?
	`, userID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[int64]struct{}{}
	for rows.Next() {
		var taskID int64
		if err := rows.Scan(&taskID); err != nil {
			return nil, err
		}
		out[taskID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (d *DB) scanIndustryProject(
	id int64,
	userID string,
	name string,
	status string,
	strategy string,
	notes string,
	paramsStr string,
	createdAt string,
	updatedAt string,
) IndustryProject {
	out := IndustryProject{
		ID:        id,
		UserID:    userID,
		Name:      name,
		Status:    status,
		Strategy:  strategy,
		Notes:     notes,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	out.Params = json.RawMessage(normalizeJSONRaw(json.RawMessage(paramsStr), "{}"))
	return out
}

func (d *DB) CreateIndustryProjectForUser(userID string, in IndustryProjectCreateInput) (IndustryProject, error) {
	userID = normalizeUserID(userID)
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return IndustryProject{}, fmt.Errorf("name is required")
	}

	now := time.Now().UTC().Format(time.RFC3339)
	status := defaultIndustryProjectStatus(in.Status)
	strategy := normalizeIndustryStrategy(in.Strategy)
	notes := strings.TrimSpace(in.Notes)
	paramsJSON := marshalJSONOrDefault(in.Params, "{}")

	res, err := d.sql.Exec(`
		INSERT INTO industry_projects (
			user_id, name, status, strategy, notes, params_json, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, name, status, strategy, notes, paramsJSON, now, now)
	if err != nil {
		return IndustryProject{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return IndustryProject{}, err
	}
	return IndustryProject{
		ID:        id,
		UserID:    userID,
		Name:      name,
		Status:    status,
		Strategy:  strategy,
		Notes:     notes,
		Params:    json.RawMessage(paramsJSON),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (d *DB) GetIndustryProjectForUser(userID string, projectID int64) (*IndustryProject, error) {
	userID = normalizeUserID(userID)
	if projectID <= 0 {
		return nil, sql.ErrNoRows
	}

	var (
		id        int64
		dbUserID  string
		name      string
		status    string
		strategy  string
		notes     string
		paramsStr string
		createdAt string
		updatedAt string
	)
	err := d.sql.QueryRow(`
		SELECT id, user_id, name, status, strategy, notes, params_json, created_at, updated_at
		  FROM industry_projects
		 WHERE user_id = ? AND id = ?
	`, userID, projectID).Scan(
		&id, &dbUserID, &name, &status, &strategy, &notes, &paramsStr, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	project := d.scanIndustryProject(id, dbUserID, name, status, strategy, notes, paramsStr, createdAt, updatedAt)
	return &project, nil
}

func (d *DB) ListIndustryProjectsForUser(userID, status string, limit int) ([]IndustryProject, error) {
	userID = normalizeUserID(userID)
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	status = normalizeIndustryProjectStatus(status)
	var (
		rows *sql.Rows
		err  error
	)
	if status == "" {
		rows, err = d.sql.Query(`
			SELECT id, user_id, name, status, strategy, notes, params_json, created_at, updated_at
			  FROM industry_projects
			 WHERE user_id = ?
			 ORDER BY updated_at DESC, id DESC
			 LIMIT ?
		`, userID, limit)
	} else {
		rows, err = d.sql.Query(`
			SELECT id, user_id, name, status, strategy, notes, params_json, created_at, updated_at
			  FROM industry_projects
			 WHERE user_id = ? AND status = ?
			 ORDER BY updated_at DESC, id DESC
			 LIMIT ?
		`, userID, status, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]IndustryProject, 0, limit)
	for rows.Next() {
		var (
			id        int64
			dbUserID  string
			name      string
			dbStatus  string
			strategy  string
			notes     string
			paramsStr string
			createdAt string
			updatedAt string
		)
		if err := rows.Scan(
			&id, &dbUserID, &name, &dbStatus, &strategy, &notes, &paramsStr, &createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, d.scanIndustryProject(
			id, dbUserID, name, dbStatus, strategy, notes, paramsStr, createdAt, updatedAt,
		))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if projects == nil {
		return []IndustryProject{}, nil
	}
	return projects, nil
}

func (d *DB) PreviewIndustryPlanForUser(userID string, projectID int64, patch IndustryPlanPatch) (IndustryPlanPreview, error) {
	userID = normalizeUserID(userID)
	if projectID <= 0 {
		return IndustryPlanPreview{}, fmt.Errorf("project_id must be positive")
	}

	project, err := d.GetIndustryProjectForUser(userID, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return IndustryPlanPreview{}, sql.ErrNoRows
		}
		return IndustryPlanPreview{}, err
	}

	nowTime := time.Now().UTC()
	now := nowTime.Format(time.RFC3339)

	inputTaskIndexToID := map[int64]int64{}
	taskRecords := make([]industryPlanTaskRecord, 0, len(patch.Tasks))
	for idx, t := range patch.Tasks {
		name := strings.TrimSpace(t.Name)
		if name == "" {
			continue
		}
		inputIndex := int64(idx + 1)
		syntheticID := inputIndex
		if !patch.Replace {
			syntheticID = 1_000_000_000 + inputIndex
		}
		inputTaskIndexToID[inputIndex] = syntheticID
		taskRecords = append(taskRecords, industryPlanTaskRecord{
			InputIndex:      inputIndex,
			SourceTaskID:    t.TaskID,
			ID:              syntheticID,
			ParentTaskID:    t.ParentTaskID,
			Name:            name,
			Activity:        normalizeIndustryActivity(t.Activity),
			TargetRuns:      t.TargetRuns,
			PlannedStart:    strings.TrimSpace(t.PlannedStart),
			PlannedEnd:      strings.TrimSpace(t.PlannedEnd),
			Priority:        t.Priority,
			ConstraintsJSON: normalizeJSONRaw(t.Constraints, "{}"),
		})
	}

	taskParents := map[int64]int64{}
	taskPlannedEnd := map[int64]time.Time{}
	existingTaskIDs := map[int64]struct{}{}
	if !patch.Replace {
		if err := d.loadIndustryTaskSchedulingMaps(userID, projectID, taskParents, taskPlannedEnd); err != nil {
			return IndustryPlanPreview{}, err
		}
		for taskID := range taskParents {
			existingTaskIDs[taskID] = struct{}{}
		}
	}
	ambiguousTaskParentRefs := 0
	sourceTaskIDToID := buildIndustrySourceTaskIDMap(taskRecords)
	for i := range taskRecords {
		parentID, ambiguous := remapIndustryTaskReference(
			taskRecords[i].ParentTaskID,
			inputTaskIndexToID,
			sourceTaskIDToID,
			len(patch.Tasks),
			patch.Replace,
			existingTaskIDs,
		)
		taskRecords[i].ParentTaskID = parentID
		if ambiguous {
			ambiguousTaskParentRefs++
		}
	}
	for _, rec := range taskRecords {
		taskParents[rec.ID] = rec.ParentTaskID
		if t, ok := parseRFC3339UTC(rec.PlannedEnd); ok {
			taskPlannedEnd[rec.ID] = t
		}
	}

	jobsForPreview := make([]IndustryJobPlanInput, len(patch.Jobs))
	copy(jobsForPreview, patch.Jobs)
	ambiguousJobTaskRefs := remapIndustryJobTaskIDs(
		jobsForPreview,
		inputTaskIndexToID,
		sourceTaskIDToID,
		len(patch.Tasks),
		patch.Replace,
		existingTaskIDs,
	)

	schedulerCfg := normalizeIndustrySchedulerInput(patch.Scheduler, project.Strategy)
	jobsSplitFrom := len(jobsForPreview)
	schedulerApplied := false
	if schedulerCfg.Enabled {
		if len(jobsForPreview) == 0 && len(taskRecords) > 0 {
			jobsForPreview = deriveIndustryJobsFromTaskRecords(taskRecords, schedulerCfg.QueueStatus)
			jobsSplitFrom = len(jobsForPreview)
		}
		if len(jobsForPreview) > 0 {
			jobsForPreview = splitAndScheduleIndustryJobs(jobsForPreview, schedulerCfg, nowTime, taskParents, taskPlannedEnd)
			schedulerApplied = true
		}
	}

	warnings := make([]string, 0, 4)
	if patch.StrictBPBypass {
		warnings = append(warnings, "strict blueprint gate bypass requested")
	}
	if schedulerCfg.Enabled && len(jobsForPreview) == 0 {
		warnings = append(warnings, "scheduler enabled but no jobs produced")
	}
	if !schedulerCfg.Enabled && len(patch.Jobs) == 0 && len(taskRecords) > 0 {
		warnings = append(warnings, "tasks provided without jobs and scheduler is disabled")
	}
	if ambiguousTaskParentRefs > 0 || ambiguousJobTaskRefs > 0 {
		warnings = append(
			warnings,
			fmt.Sprintf(
				"%s (task_parents=%d jobs=%d)",
				industryTaskRefAmbiguityWarning,
				ambiguousTaskParentRefs,
				ambiguousJobTaskRefs,
			),
		)
	}

	blueprintPool := []IndustryBlueprintPool{}
	taskRecordsForCaps := make([]industryPlanTaskRecord, 0, len(taskRecords)+32)
	taskRecordsForCaps = append(taskRecordsForCaps, taskRecords...)
	if !patch.Replace {
		existingTasks, loadErr := d.listIndustryTaskRecordsForProject(userID, projectID)
		if loadErr != nil {
			return IndustryPlanPreview{}, loadErr
		}
		taskRecordsForCaps = append(taskRecordsForCaps, existingTasks...)
	}
	if !patch.Replace {
		blueprintPool, err = d.listIndustryBlueprintPoolForProject(userID, projectID)
		if err != nil {
			return IndustryPlanPreview{}, err
		}
	}
	blueprintPool = mergeIndustryBlueprintPool(blueprintPool, patch.Blueprints)
	if len(jobsForPreview) > 0 && len(taskRecordsForCaps) > 0 {
		var capWarnings []string
		jobsForPreview, capWarnings = applyBlueprintRunCapsToJobs(
			jobsForPreview,
			taskRecordsForCaps,
			blueprintPool,
			!patch.StrictBPBypass,
		)
		if len(capWarnings) > 0 {
			warnings = append(warnings, capWarnings...)
		}
	}

	materialCount := 0
	for _, m := range patch.Materials {
		if m.TypeID > 0 {
			materialCount++
		}
	}
	blueprintCount := 0
	for _, bp := range patch.Blueprints {
		if bp.BlueprintTypeID > 0 {
			blueprintCount++
		}
	}

	projectStatus := project.Status
	if normalized := normalizeIndustryProjectStatus(patch.ProjectStatus); normalized != "" {
		projectStatus = normalized
	}

	summary := IndustryPlanSummary{
		ProjectID:        projectID,
		ProjectStatus:    projectStatus,
		Replaced:         patch.Replace,
		TasksInserted:    len(taskRecords),
		JobsInserted:     len(jobsForPreview),
		MaterialsUpsert:  materialCount,
		BlueprintsUpsert: blueprintCount,
		SchedulerApplied: schedulerApplied,
		JobsSplitFrom:    jobsSplitFrom,
		JobsPlannedTotal: len(jobsForPreview),
		Warnings:         warnings,
		UpdatedAt:        now,
	}

	taskPreview := make([]IndustryTaskPreview, 0, len(taskRecords))
	sort.SliceStable(taskRecords, func(i, j int) bool {
		return taskRecords[i].InputIndex < taskRecords[j].InputIndex
	})
	for _, rec := range taskRecords {
		taskPreview = append(taskPreview, IndustryTaskPreview{
			InputIndex:   rec.InputIndex,
			TaskID:       rec.ID,
			ParentTaskID: rec.ParentTaskID,
			Name:         rec.Name,
			Activity:     rec.Activity,
			TargetRuns:   rec.TargetRuns,
			PlannedStart: rec.PlannedStart,
			PlannedEnd:   rec.PlannedEnd,
			Priority:     rec.Priority,
		})
	}

	return IndustryPlanPreview{
		ProjectID: projectID,
		Replace:   patch.Replace,
		Summary:   summary,
		Tasks:     taskPreview,
		Jobs:      jobsForPreview,
		Warnings:  warnings,
	}, nil
}

func (d *DB) ApplyIndustryPlanForUser(userID string, projectID int64, patch IndustryPlanPatch) (IndustryPlanSummary, error) {
	userID = normalizeUserID(userID)
	if projectID <= 0 {
		return IndustryPlanSummary{}, fmt.Errorf("project_id must be positive")
	}

	project, err := d.GetIndustryProjectForUser(userID, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return IndustryPlanSummary{}, sql.ErrNoRows
		}
		return IndustryPlanSummary{}, err
	}

	tx, err := d.sql.Begin()
	if err != nil {
		return IndustryPlanSummary{}, err
	}
	defer tx.Rollback()

	if patch.Replace {
		if _, err := tx.Exec(`DELETE FROM industry_jobs WHERE user_id = ? AND project_id = ?`, userID, projectID); err != nil {
			return IndustryPlanSummary{}, err
		}
		if _, err := tx.Exec(`DELETE FROM industry_tasks WHERE user_id = ? AND project_id = ?`, userID, projectID); err != nil {
			return IndustryPlanSummary{}, err
		}
		if _, err := tx.Exec(`DELETE FROM industry_material_plan WHERE user_id = ? AND project_id = ?`, userID, projectID); err != nil {
			return IndustryPlanSummary{}, err
		}
	}
	if patch.Replace || patch.ReplaceBlueprintPool {
		if _, err := tx.Exec(`DELETE FROM industry_blueprint_pool WHERE user_id = ? AND project_id = ?`, userID, projectID); err != nil {
			return IndustryPlanSummary{}, err
		}
	}

	nowTime := time.Now().UTC()
	now := nowTime.Format(time.RFC3339)
	taskInsertCount := 0
	insertedTaskRecords := make([]industryPlanTaskRecord, 0, len(patch.Tasks))
	inputTaskIndexToID := map[int64]int64{}
	existingTaskIDs := map[int64]struct{}{}
	if !patch.Replace {
		existingTaskIDs, err = loadIndustryTaskIDSetTx(tx, userID, projectID)
		if err != nil {
			return IndustryPlanSummary{}, err
		}
	}
	if len(patch.Tasks) > 0 {
		stmt, err := tx.Prepare(`
			INSERT INTO industry_tasks (
				user_id, project_id, parent_task_id, name, activity, product_type_id, target_runs,
				planned_start, planned_end, priority, status, constraints_json, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			return IndustryPlanSummary{}, err
		}
		defer stmt.Close()

		for idx, t := range patch.Tasks {
			name := strings.TrimSpace(t.Name)
			if name == "" {
				continue
			}
			constraintsJSON := normalizeJSONRaw(t.Constraints, "{}")
			res, err := stmt.Exec(
				userID,
				projectID,
				nil,
				name,
				normalizeIndustryActivity(t.Activity),
				t.ProductTypeID,
				t.TargetRuns,
				strings.TrimSpace(t.PlannedStart),
				strings.TrimSpace(t.PlannedEnd),
				t.Priority,
				defaultIndustryTaskStatus(t.Status),
				constraintsJSON,
				now,
				now,
			)
			if err != nil {
				return IndustryPlanSummary{}, err
			}
			insertedID, err := res.LastInsertId()
			if err != nil {
				return IndustryPlanSummary{}, err
			}
			inputIndex := int64(idx + 1)
			inputTaskIndexToID[inputIndex] = insertedID
			insertedTaskRecords = append(insertedTaskRecords, industryPlanTaskRecord{
				InputIndex:      inputIndex,
				SourceTaskID:    t.TaskID,
				ID:              insertedID,
				ParentTaskID:    t.ParentTaskID,
				Name:            name,
				Activity:        normalizeIndustryActivity(t.Activity),
				TargetRuns:      t.TargetRuns,
				PlannedStart:    strings.TrimSpace(t.PlannedStart),
				PlannedEnd:      strings.TrimSpace(t.PlannedEnd),
				Priority:        t.Priority,
				ConstraintsJSON: constraintsJSON,
			})
			taskInsertCount++
		}
	}

	ambiguousTaskParentRefs := 0
	sourceTaskIDToID := buildIndustrySourceTaskIDMap(insertedTaskRecords)
	if len(insertedTaskRecords) > 0 {
		for i := range insertedTaskRecords {
			parentID := insertedTaskRecords[i].ParentTaskID
			parentID, ambiguous := remapIndustryTaskReference(
				parentID,
				inputTaskIndexToID,
				sourceTaskIDToID,
				len(patch.Tasks),
				patch.Replace,
				existingTaskIDs,
			)
			if ambiguous {
				ambiguousTaskParentRefs++
			}
			insertedTaskRecords[i].ParentTaskID = parentID
			if _, err := tx.Exec(`
				UPDATE industry_tasks
				   SET parent_task_id = ?, updated_at = ?
				 WHERE user_id = ? AND project_id = ? AND id = ?
			`, nullablePositiveInt64(parentID), now, userID, projectID, insertedTaskRecords[i].ID); err != nil {
				return IndustryPlanSummary{}, err
			}
		}
	}

	schedulerCfg := normalizeIndustrySchedulerInput(patch.Scheduler, project.Strategy)
	jobsForInsert := make([]IndustryJobPlanInput, len(patch.Jobs))
	copy(jobsForInsert, patch.Jobs)
	ambiguousJobTaskRefs := remapIndustryJobTaskIDs(
		jobsForInsert,
		inputTaskIndexToID,
		sourceTaskIDToID,
		len(patch.Tasks),
		patch.Replace,
		existingTaskIDs,
	)
	jobsSplitFrom := len(jobsForInsert)
	schedulerApplied := false
	if schedulerCfg.Enabled {
		if len(jobsForInsert) == 0 && len(insertedTaskRecords) > 0 {
			jobsForInsert = deriveIndustryJobsFromTaskRecords(insertedTaskRecords, schedulerCfg.QueueStatus)
			jobsSplitFrom = len(jobsForInsert)
		}
		if len(jobsForInsert) > 0 {
			taskParents := map[int64]int64{}
			taskPlannedEnd := map[int64]time.Time{}
			if err := loadIndustryTaskSchedulingMapsTx(tx, userID, projectID, taskParents, taskPlannedEnd); err != nil {
				return IndustryPlanSummary{}, err
			}
			jobsForInsert = splitAndScheduleIndustryJobs(jobsForInsert, schedulerCfg, nowTime, taskParents, taskPlannedEnd)
			schedulerApplied = true
		}
	}

	warnings := make([]string, 0, 4)
	if patch.StrictBPBypass {
		warnings = append(warnings, "strict blueprint gate bypass requested")
	}
	if schedulerCfg.Enabled && len(jobsForInsert) == 0 {
		warnings = append(warnings, "scheduler enabled but no jobs produced")
	}
	if !schedulerCfg.Enabled && len(patch.Jobs) == 0 && len(insertedTaskRecords) > 0 {
		warnings = append(warnings, "tasks provided without jobs and scheduler is disabled")
	}
	if ambiguousTaskParentRefs > 0 || ambiguousJobTaskRefs > 0 {
		warnings = append(
			warnings,
			fmt.Sprintf(
				"%s (task_parents=%d jobs=%d)",
				industryTaskRefAmbiguityWarning,
				ambiguousTaskParentRefs,
				ambiguousJobTaskRefs,
			),
		)
	}

	blueprintPool := []IndustryBlueprintPool{}
	if !patch.Replace {
		blueprintPool, err = listIndustryBlueprintPoolForProjectTx(tx, userID, projectID)
		if err != nil {
			return IndustryPlanSummary{}, err
		}
	}
	blueprintPool = mergeIndustryBlueprintPool(blueprintPool, patch.Blueprints)
	taskRecordsForCaps, err := listIndustryTaskRecordsForProjectTx(tx, userID, projectID)
	if err != nil {
		return IndustryPlanSummary{}, err
	}
	if len(jobsForInsert) > 0 && len(taskRecordsForCaps) > 0 {
		var capWarnings []string
		jobsForInsert, capWarnings = applyBlueprintRunCapsToJobs(
			jobsForInsert,
			taskRecordsForCaps,
			blueprintPool,
			!patch.StrictBPBypass,
		)
		if len(capWarnings) > 0 {
			warnings = append(warnings, capWarnings...)
		}
	}

	jobInsertCount := 0
	if len(jobsForInsert) > 0 {
		stmt, err := tx.Prepare(`
			INSERT INTO industry_jobs (
				user_id, project_id, task_id, character_id, facility_id, activity, runs,
				duration_seconds, cost_isk, status, started_at, finished_at, external_job_id, notes, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			return IndustryPlanSummary{}, err
		}
		defer stmt.Close()

		for _, j := range jobsForInsert {
			if _, err := stmt.Exec(
				userID,
				projectID,
				nullablePositiveInt64(j.TaskID),
				j.CharacterID,
				j.FacilityID,
				normalizeIndustryActivity(j.Activity),
				j.Runs,
				j.DurationSeconds,
				j.CostISK,
				defaultIndustryJobStatus(j.Status),
				strings.TrimSpace(j.StartedAt),
				strings.TrimSpace(j.FinishedAt),
				j.ExternalJobID,
				strings.TrimSpace(j.Notes),
				now,
				now,
			); err != nil {
				return IndustryPlanSummary{}, err
			}
			jobInsertCount++
		}
	}

	materialUpsertCount := 0
	if len(patch.Materials) > 0 {
		stmt, err := tx.Prepare(`
			INSERT INTO industry_material_plan (
				user_id, project_id, task_id, type_id, type_name, required_qty, available_qty,
				buy_qty, build_qty, unit_cost_isk, source, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(user_id, project_id, task_id, type_id)
			DO UPDATE SET
				type_name = excluded.type_name,
				required_qty = excluded.required_qty,
				available_qty = excluded.available_qty,
				buy_qty = excluded.buy_qty,
				build_qty = excluded.build_qty,
				unit_cost_isk = excluded.unit_cost_isk,
				source = excluded.source,
				updated_at = excluded.updated_at
		`)
		if err != nil {
			return IndustryPlanSummary{}, err
		}
		defer stmt.Close()

		for _, m := range patch.Materials {
			if m.TypeID <= 0 {
				continue
			}
			if _, err := stmt.Exec(
				userID,
				projectID,
				m.TaskID,
				m.TypeID,
				strings.TrimSpace(m.TypeName),
				m.RequiredQty,
				m.AvailableQty,
				m.BuyQty,
				m.BuildQty,
				m.UnitCostISK,
				normalizeIndustryMaterialSource(m.Source),
				now,
			); err != nil {
				return IndustryPlanSummary{}, err
			}
			materialUpsertCount++
		}
	}

	blueprintUpsertCount := 0
	if len(patch.Blueprints) > 0 {
		stmt, err := tx.Prepare(`
			INSERT INTO industry_blueprint_pool (
				user_id, project_id, blueprint_type_id, blueprint_name, location_id,
				quantity, me, te, is_bpo, available_runs, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(user_id, project_id, blueprint_type_id, location_id, is_bpo)
			DO UPDATE SET
				blueprint_name = excluded.blueprint_name,
				quantity = excluded.quantity,
				me = excluded.me,
				te = excluded.te,
				available_runs = excluded.available_runs,
				updated_at = excluded.updated_at
		`)
		if err != nil {
			return IndustryPlanSummary{}, err
		}
		defer stmt.Close()

		for _, bp := range patch.Blueprints {
			if bp.BlueprintTypeID <= 0 {
				continue
			}
			isBPO := 0
			if bp.IsBPO {
				isBPO = 1
			}
			if _, err := stmt.Exec(
				userID,
				projectID,
				bp.BlueprintTypeID,
				strings.TrimSpace(bp.BlueprintName),
				bp.LocationID,
				bp.Quantity,
				bp.ME,
				bp.TE,
				isBPO,
				normalizeIndustryBlueprintAvailableRuns(bp.IsBPO, bp.AvailableRuns),
				now,
			); err != nil {
				return IndustryPlanSummary{}, err
			}
			blueprintUpsertCount++
		}
	}

	projectStatus := project.Status
	if normalized := normalizeIndustryProjectStatus(patch.ProjectStatus); normalized != "" {
		projectStatus = normalized
	}
	if _, err := tx.Exec(`
		UPDATE industry_projects
		   SET status = ?, updated_at = ?
		 WHERE user_id = ? AND id = ?
	`, projectStatus, now, userID, projectID); err != nil {
		return IndustryPlanSummary{}, err
	}

	if err := tx.Commit(); err != nil {
		return IndustryPlanSummary{}, err
	}

	return IndustryPlanSummary{
		ProjectID:        projectID,
		ProjectStatus:    projectStatus,
		Replaced:         patch.Replace,
		TasksInserted:    taskInsertCount,
		JobsInserted:     jobInsertCount,
		MaterialsUpsert:  materialUpsertCount,
		BlueprintsUpsert: blueprintUpsertCount,
		SchedulerApplied: schedulerApplied,
		JobsSplitFrom:    jobsSplitFrom,
		JobsPlannedTotal: len(jobsForInsert),
		Warnings:         warnings,
		UpdatedAt:        now,
	}, nil
}

func (d *DB) getIndustryJobForUser(userID string, jobID int64) (*IndustryJob, error) {
	var (
		j IndustryJob
	)
	err := d.sql.QueryRow(`
		SELECT id, user_id, project_id, COALESCE(task_id, 0), character_id, facility_id, activity, runs,
		       duration_seconds, cost_isk, status, started_at, finished_at, external_job_id, notes, created_at, updated_at
		  FROM industry_jobs
		 WHERE user_id = ? AND id = ?
	`, userID, jobID).Scan(
		&j.ID,
		&j.UserID,
		&j.ProjectID,
		&j.TaskID,
		&j.CharacterID,
		&j.FacilityID,
		&j.Activity,
		&j.Runs,
		&j.DurationSeconds,
		&j.CostISK,
		&j.Status,
		&j.StartedAt,
		&j.FinishedAt,
		&j.ExternalJobID,
		&j.Notes,
		&j.CreatedAt,
		&j.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func (d *DB) getIndustryTaskForUser(userID string, taskID int64) (*IndustryTask, error) {
	var (
		t              IndustryTask
		constraintsRaw string
	)
	err := d.sql.QueryRow(`
		SELECT id, user_id, project_id, COALESCE(parent_task_id, 0), name, activity, product_type_id, target_runs,
		       planned_start, planned_end, priority, status, constraints_json, created_at, updated_at
		  FROM industry_tasks
		 WHERE user_id = ? AND id = ?
	`, userID, taskID).Scan(
		&t.ID,
		&t.UserID,
		&t.ProjectID,
		&t.ParentTaskID,
		&t.Name,
		&t.Activity,
		&t.ProductTypeID,
		&t.TargetRuns,
		&t.PlannedStart,
		&t.PlannedEnd,
		&t.Priority,
		&t.Status,
		&constraintsRaw,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	t.Constraints = json.RawMessage(normalizeJSONRaw(json.RawMessage(constraintsRaw), "{}"))
	return &t, nil
}

func (d *DB) UpdateIndustryTaskStatusForUser(
	userID string,
	taskID int64,
	status string,
	priority *int,
) (*IndustryTask, error) {
	userID = normalizeUserID(userID)
	if taskID <= 0 {
		return nil, fmt.Errorf("task_id must be positive")
	}
	normalizedStatus := normalizeIndustryTaskStatus(status)
	if normalizedStatus == "" {
		return nil, fmt.Errorf("invalid task status")
	}

	task, err := d.getIndustryTaskForUser(userID, taskID)
	if err != nil {
		return nil, err
	}
	nextPriority := task.Priority
	if priority != nil {
		nextPriority = *priority
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err = d.sql.Exec(`
		UPDATE industry_tasks
		   SET status = ?, priority = ?, updated_at = ?
		 WHERE user_id = ? AND id = ?
	`, normalizedStatus, nextPriority, now, userID, taskID)
	if err != nil {
		return nil, err
	}
	return d.getIndustryTaskForUser(userID, taskID)
}

func (d *DB) UpdateIndustryTaskStatusesForUser(
	userID string,
	taskIDs []int64,
	status string,
	priority *int,
) ([]IndustryTask, error) {
	userID = normalizeUserID(userID)
	if len(taskIDs) == 0 {
		return nil, fmt.Errorf("task_ids are required")
	}
	normalizedStatus := normalizeIndustryTaskStatus(status)
	if normalizedStatus == "" {
		return nil, fmt.Errorf("invalid task status")
	}

	uniqueIDs := make([]int64, 0, len(taskIDs))
	seen := make(map[int64]struct{}, len(taskIDs))
	for _, taskID := range taskIDs {
		if taskID <= 0 {
			return nil, fmt.Errorf("task_id must be positive")
		}
		if _, ok := seen[taskID]; ok {
			continue
		}
		seen[taskID] = struct{}{}
		uniqueIDs = append(uniqueIDs, taskID)
	}
	if len(uniqueIDs) == 0 {
		return nil, fmt.Errorf("task_ids are required")
	}

	tx, err := d.sql.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	placeholders := strings.Repeat("?,", len(uniqueIDs))
	placeholders = strings.TrimSuffix(placeholders, ",")
	query := `
		SELECT id, user_id, project_id, COALESCE(parent_task_id, 0), name, activity, product_type_id, target_runs,
		       planned_start, planned_end, priority, status, constraints_json, created_at, updated_at
		  FROM industry_tasks
		 WHERE user_id = ? AND id IN (` + placeholders + `)
	`
	args := make([]interface{}, 0, len(uniqueIDs)+1)
	args = append(args, userID)
	for _, id := range uniqueIDs {
		args = append(args, id)
	}

	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasksByID := make(map[int64]IndustryTask, len(uniqueIDs))
	for rows.Next() {
		var (
			task           IndustryTask
			constraintsRaw string
		)
		if err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.ProjectID,
			&task.ParentTaskID,
			&task.Name,
			&task.Activity,
			&task.ProductTypeID,
			&task.TargetRuns,
			&task.PlannedStart,
			&task.PlannedEnd,
			&task.Priority,
			&task.Status,
			&constraintsRaw,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, err
		}
		task.Constraints = json.RawMessage(normalizeJSONRaw(json.RawMessage(constraintsRaw), "{}"))
		tasksByID[task.ID] = task
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(tasksByID) != len(uniqueIDs) {
		return nil, sql.ErrNoRows
	}

	now := time.Now().UTC().Format(time.RFC3339)
	stmt, err := tx.Prepare(`
		UPDATE industry_tasks
		   SET status = ?, priority = ?, updated_at = ?
		 WHERE user_id = ? AND id = ?
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	updated := make([]IndustryTask, 0, len(uniqueIDs))
	for _, taskID := range uniqueIDs {
		task := tasksByID[taskID]
		nextPriority := task.Priority
		if priority != nil {
			nextPriority = *priority
		}
		if _, err := stmt.Exec(normalizedStatus, nextPriority, now, userID, taskID); err != nil {
			return nil, err
		}
		task.Status = normalizedStatus
		task.Priority = nextPriority
		task.UpdatedAt = now
		updated = append(updated, task)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return updated, nil
}

func (d *DB) UpdateIndustryTaskPrioritiesForUser(
	userID string,
	taskIDs []int64,
	priority int,
) ([]IndustryTask, error) {
	userID = normalizeUserID(userID)
	if len(taskIDs) == 0 {
		return nil, fmt.Errorf("task_ids are required")
	}

	uniqueIDs := make([]int64, 0, len(taskIDs))
	seen := make(map[int64]struct{}, len(taskIDs))
	for _, taskID := range taskIDs {
		if taskID <= 0 {
			return nil, fmt.Errorf("task_id must be positive")
		}
		if _, ok := seen[taskID]; ok {
			continue
		}
		seen[taskID] = struct{}{}
		uniqueIDs = append(uniqueIDs, taskID)
	}
	if len(uniqueIDs) == 0 {
		return nil, fmt.Errorf("task_ids are required")
	}

	tx, err := d.sql.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	placeholders := strings.Repeat("?,", len(uniqueIDs))
	placeholders = strings.TrimSuffix(placeholders, ",")
	query := `
		SELECT id, user_id, project_id, COALESCE(parent_task_id, 0), name, activity, product_type_id, target_runs,
		       planned_start, planned_end, priority, status, constraints_json, created_at, updated_at
		  FROM industry_tasks
		 WHERE user_id = ? AND id IN (` + placeholders + `)
	`
	args := make([]interface{}, 0, len(uniqueIDs)+1)
	args = append(args, userID)
	for _, id := range uniqueIDs {
		args = append(args, id)
	}

	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasksByID := make(map[int64]IndustryTask, len(uniqueIDs))
	for rows.Next() {
		var (
			task           IndustryTask
			constraintsRaw string
		)
		if err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.ProjectID,
			&task.ParentTaskID,
			&task.Name,
			&task.Activity,
			&task.ProductTypeID,
			&task.TargetRuns,
			&task.PlannedStart,
			&task.PlannedEnd,
			&task.Priority,
			&task.Status,
			&constraintsRaw,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, err
		}
		task.Constraints = json.RawMessage(normalizeJSONRaw(json.RawMessage(constraintsRaw), "{}"))
		tasksByID[task.ID] = task
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(tasksByID) != len(uniqueIDs) {
		return nil, sql.ErrNoRows
	}

	now := time.Now().UTC().Format(time.RFC3339)
	stmt, err := tx.Prepare(`
		UPDATE industry_tasks
		   SET priority = ?, updated_at = ?
		 WHERE user_id = ? AND id = ?
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	updated := make([]IndustryTask, 0, len(uniqueIDs))
	for _, taskID := range uniqueIDs {
		task := tasksByID[taskID]
		if _, err := stmt.Exec(priority, now, userID, taskID); err != nil {
			return nil, err
		}
		task.Priority = priority
		task.UpdatedAt = now
		updated = append(updated, task)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return updated, nil
}

func (d *DB) RebalanceIndustryProjectMaterialsFromStockForUser(
	userID string,
	projectID int64,
	stockByType map[int32]int64,
	stockByTypeLocation map[int32]map[int64]int64,
	warehouseScope string,
	strategy string,
) ([]IndustryMaterialPlan, error) {
	userID = normalizeUserID(userID)
	if projectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if _, err := d.GetIndustryProjectForUser(userID, projectID); err != nil {
		return nil, err
	}

	normalizedStrategy := normalizeIndustryMaterialRebalanceStrategy(strategy)
	normalizedWarehouseScope := normalizeIndustryMaterialWarehouseScope(warehouseScope)
	stockPools := buildIndustryMaterialStockPools(stockByType, stockByTypeLocation)

	tx, err := d.sql.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	taskPreferredLocation, err := listIndustryTaskPreferredLocationForProjectTx(tx, userID, projectID)
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(`
		SELECT id, user_id, project_id, task_id, type_id, type_name, required_qty, available_qty,
		       buy_qty, build_qty, unit_cost_isk, source, updated_at
		  FROM industry_material_plan
		 WHERE user_id = ? AND project_id = ?
		 ORDER BY type_id ASC, task_id ASC, id ASC
	`, userID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	materials := make([]IndustryMaterialPlan, 0, 128)
	for rows.Next() {
		var row IndustryMaterialPlan
		if err := rows.Scan(
			&row.ID,
			&row.UserID,
			&row.ProjectID,
			&row.TaskID,
			&row.TypeID,
			&row.TypeName,
			&row.RequiredQty,
			&row.AvailableQty,
			&row.BuyQty,
			&row.BuildQty,
			&row.UnitCostISK,
			&row.Source,
			&row.UpdatedAt,
		); err != nil {
			return nil, err
		}
		materials = append(materials, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(materials) == 0 {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return []IndustryMaterialPlan{}, nil
	}

	stmt, err := tx.Prepare(`
		UPDATE industry_material_plan
		   SET available_qty = ?, buy_qty = ?, build_qty = ?, source = ?, updated_at = ?
		 WHERE user_id = ? AND project_id = ? AND id = ?
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	updated := make([]IndustryMaterialPlan, 0, len(materials))

	for _, row := range materials {
		pool := stockPools[row.TypeID]

		nextAvailable := int64(0)
		requiredQty := row.RequiredQty
		if requiredQty < 0 {
			requiredQty = 0
		}
		preferredLocationID := taskPreferredLocation[row.TaskID]
		if normalizedWarehouseScope != "global" && preferredLocationID > 0 {
			nextAvailable += industryMaterialAllocateFromLocation(pool, preferredLocationID, requiredQty)
		}

		remainingForGlobal := requiredQty - nextAvailable
		if remainingForGlobal < 0 {
			remainingForGlobal = 0
		}
		if normalizedWarehouseScope == "strict_location" && preferredLocationID > 0 {
			remainingForGlobal = 0
		}
		nextAvailable += industryMaterialAllocateGlobal(pool, remainingForGlobal)

		remaining := row.RequiredQty - nextAvailable
		if remaining < 0 {
			remaining = 0
		}

		var nextBuy int64
		var nextBuild int64
		switch normalizedStrategy {
		case "buy":
			nextBuy = remaining
		case "build":
			nextBuild = remaining
		default:
			oldCovered := row.BuyQty + row.BuildQty
			switch {
			case remaining <= 0:
				nextBuy = 0
				nextBuild = 0
			case oldCovered > 0 && row.BuildQty <= 0:
				nextBuy = remaining
			case oldCovered > 0 && row.BuyQty <= 0:
				nextBuild = remaining
			case oldCovered > 0:
				buildRatio := float64(row.BuildQty) / float64(oldCovered)
				nextBuild = int64(math.Round(float64(remaining) * buildRatio))
				if nextBuild < 0 {
					nextBuild = 0
				}
				if nextBuild > remaining {
					nextBuild = remaining
				}
				nextBuy = remaining - nextBuild
			default:
				switch normalizeIndustryMaterialSource(row.Source) {
				case "build", "reprocess":
					nextBuild = remaining
				default:
					nextBuy = remaining
				}
			}
		}

		nextSource := normalizeIndustryMaterialSource(row.Source)
		switch {
		case nextAvailable >= row.RequiredQty:
			nextSource = "stock"
		case nextBuy == 0 && nextBuild > 0:
			if nextSource == "reprocess" {
				nextSource = "reprocess"
			} else {
				nextSource = "build"
			}
		case nextBuild == 0 && nextBuy > 0:
			if nextSource == "contract" {
				nextSource = "contract"
			} else {
				nextSource = "market"
			}
		case nextBuild > 0 && nextBuy > 0:
			if nextSource == "stock" {
				nextSource = "market"
			}
		default:
			nextSource = "stock"
		}

		if _, err := stmt.Exec(
			nextAvailable,
			nextBuy,
			nextBuild,
			nextSource,
			now,
			userID,
			projectID,
			row.ID,
		); err != nil {
			return nil, err
		}

		row.AvailableQty = nextAvailable
		row.BuyQty = nextBuy
		row.BuildQty = nextBuild
		row.Source = nextSource
		row.UpdatedAt = now
		updated = append(updated, row)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return updated, nil
}

func computeIndustryJobStatusUpdate(
	job *IndustryJob,
	normalizedStatus string,
	startedAt string,
	finishedAt string,
	notes string,
	now string,
) (string, string, string, error) {
	nextStartedAt := job.StartedAt
	nextFinishedAt := job.FinishedAt
	nextNotes := job.Notes

	if trimmed := strings.TrimSpace(startedAt); trimmed != "" {
		normalizedStartedAt, err := normalizeOptionalRFC3339(trimmed, "started_at")
		if err != nil {
			return "", "", "", err
		}
		nextStartedAt = normalizedStartedAt
	}
	if trimmed := strings.TrimSpace(finishedAt); trimmed != "" {
		normalizedFinishedAt, err := normalizeOptionalRFC3339(trimmed, "finished_at")
		if err != nil {
			return "", "", "", err
		}
		nextFinishedAt = normalizedFinishedAt
	}
	if trimmed := strings.TrimSpace(notes); trimmed != "" {
		nextNotes = trimmed
	}

	if !jobTerminalStatus(normalizedStatus) {
		nextFinishedAt = ""
	}
	if normalizedStatus == IndustryJobStatusActive && strings.TrimSpace(nextStartedAt) == "" {
		nextStartedAt = now
	}
	if (normalizedStatus == IndustryJobStatusCompleted || normalizedStatus == IndustryJobStatusFailed || normalizedStatus == IndustryJobStatusCancelled) &&
		strings.TrimSpace(nextFinishedAt) == "" {
		nextFinishedAt = now
	}
	return nextStartedAt, nextFinishedAt, nextNotes, nil
}

func (d *DB) UpdateIndustryJobStatusForUser(
	userID string,
	jobID int64,
	status string,
	startedAt string,
	finishedAt string,
	notes string,
) (*IndustryJob, error) {
	userID = normalizeUserID(userID)
	if jobID <= 0 {
		return nil, fmt.Errorf("job_id must be positive")
	}
	normalizedStatus := normalizeIndustryJobStatus(status)
	if normalizedStatus == "" {
		return nil, fmt.Errorf("invalid job status")
	}

	job, err := d.getIndustryJobForUser(userID, jobID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	nextStartedAt, nextFinishedAt, nextNotes, err := computeIndustryJobStatusUpdate(
		job, normalizedStatus, startedAt, finishedAt, notes, now,
	)
	if err != nil {
		return nil, err
	}

	_, err = d.sql.Exec(`
		UPDATE industry_jobs
		   SET status = ?, started_at = ?, finished_at = ?, notes = ?, updated_at = ?
		 WHERE user_id = ? AND id = ?
	`, normalizedStatus, nextStartedAt, nextFinishedAt, nextNotes, now, userID, jobID)
	if err != nil {
		return nil, err
	}
	return d.getIndustryJobForUser(userID, jobID)
}

func (d *DB) UpdateIndustryJobStatusesForUser(
	userID string,
	jobIDs []int64,
	status string,
	startedAt string,
	finishedAt string,
	notes string,
) ([]IndustryJob, error) {
	userID = normalizeUserID(userID)
	if len(jobIDs) == 0 {
		return nil, fmt.Errorf("job_ids are required")
	}
	normalizedStatus := normalizeIndustryJobStatus(status)
	if normalizedStatus == "" {
		return nil, fmt.Errorf("invalid job status")
	}

	uniqueIDs := make([]int64, 0, len(jobIDs))
	seen := make(map[int64]struct{}, len(jobIDs))
	for _, jobID := range jobIDs {
		if jobID <= 0 {
			return nil, fmt.Errorf("job_id must be positive")
		}
		if _, ok := seen[jobID]; ok {
			continue
		}
		seen[jobID] = struct{}{}
		uniqueIDs = append(uniqueIDs, jobID)
	}
	if len(uniqueIDs) == 0 {
		return nil, fmt.Errorf("job_ids are required")
	}

	tx, err := d.sql.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	placeholders := strings.Repeat("?,", len(uniqueIDs))
	placeholders = strings.TrimSuffix(placeholders, ",")
	query := `
		SELECT id, user_id, project_id, COALESCE(task_id, 0), character_id, facility_id, activity, runs,
		       duration_seconds, cost_isk, status, started_at, finished_at, external_job_id, notes, created_at, updated_at
		  FROM industry_jobs
		 WHERE user_id = ? AND id IN (` + placeholders + `)
	`
	args := make([]interface{}, 0, len(uniqueIDs)+1)
	args = append(args, userID)
	for _, id := range uniqueIDs {
		args = append(args, id)
	}

	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobsByID := make(map[int64]IndustryJob, len(uniqueIDs))
	for rows.Next() {
		var j IndustryJob
		if err := rows.Scan(
			&j.ID,
			&j.UserID,
			&j.ProjectID,
			&j.TaskID,
			&j.CharacterID,
			&j.FacilityID,
			&j.Activity,
			&j.Runs,
			&j.DurationSeconds,
			&j.CostISK,
			&j.Status,
			&j.StartedAt,
			&j.FinishedAt,
			&j.ExternalJobID,
			&j.Notes,
			&j.CreatedAt,
			&j.UpdatedAt,
		); err != nil {
			return nil, err
		}
		jobsByID[j.ID] = j
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(jobsByID) != len(uniqueIDs) {
		return nil, sql.ErrNoRows
	}

	now := time.Now().UTC().Format(time.RFC3339)
	stmt, err := tx.Prepare(`
		UPDATE industry_jobs
		   SET status = ?, started_at = ?, finished_at = ?, notes = ?, updated_at = ?
		 WHERE user_id = ? AND id = ?
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	updated := make([]IndustryJob, 0, len(uniqueIDs))
	for _, jobID := range uniqueIDs {
		job := jobsByID[jobID]
		nextStartedAt, nextFinishedAt, nextNotes, updateErr := computeIndustryJobStatusUpdate(
			&job, normalizedStatus, startedAt, finishedAt, notes, now,
		)
		if updateErr != nil {
			return nil, updateErr
		}
		if _, err := stmt.Exec(
			normalizedStatus,
			nextStartedAt,
			nextFinishedAt,
			nextNotes,
			now,
			userID,
			jobID,
		); err != nil {
			return nil, err
		}

		job.Status = normalizedStatus
		job.StartedAt = nextStartedAt
		job.FinishedAt = nextFinishedAt
		job.Notes = nextNotes
		job.UpdatedAt = now
		updated = append(updated, job)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return updated, nil
}

func (d *DB) GetIndustryLedgerForUser(userID string, opt IndustryLedgerOptions) (IndustryLedger, error) {
	userID = normalizeUserID(userID)
	limit := opt.Limit
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}
	status := normalizeIndustryJobStatus(opt.Status)

	args := []interface{}{userID}
	query := `
		SELECT
			j.id,
			j.project_id,
			COALESCE(p.name, ''),
			COALESCE(j.task_id, 0),
			COALESCE(t.name, ''),
			j.character_id,
			j.facility_id,
			j.activity,
			j.runs,
			j.duration_seconds,
			j.cost_isk,
			j.status,
			j.started_at,
			j.finished_at,
			j.external_job_id,
			j.notes,
			j.updated_at
		FROM industry_jobs j
		INNER JOIN industry_projects p
			ON p.id = j.project_id AND p.user_id = j.user_id
		LEFT JOIN industry_tasks t
			ON t.id = j.task_id AND t.user_id = j.user_id
		WHERE j.user_id = ?
	`
	if opt.ProjectID > 0 {
		query += " AND j.project_id = ?"
		args = append(args, opt.ProjectID)
	}
	if status != "" {
		query += " AND j.status = ?"
		args = append(args, status)
	}
	query += " ORDER BY j.updated_at DESC, j.id DESC LIMIT ?"
	args = append(args, limit)

	rows, err := d.sql.Query(query, args...)
	if err != nil {
		return IndustryLedger{}, err
	}
	defer rows.Close()

	entries := make([]IndustryLedgerEntry, 0, limit)
	for rows.Next() {
		var e IndustryLedgerEntry
		if err := rows.Scan(
			&e.JobID,
			&e.ProjectID,
			&e.ProjectName,
			&e.TaskID,
			&e.TaskName,
			&e.CharacterID,
			&e.FacilityID,
			&e.Activity,
			&e.Runs,
			&e.DurationSeconds,
			&e.CostISK,
			&e.Status,
			&e.StartedAt,
			&e.FinishedAt,
			&e.ExternalJobID,
			&e.Notes,
			&e.LastUpdatedAtUTC,
		); err != nil {
			return IndustryLedger{}, err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return IndustryLedger{}, err
	}
	if entries == nil {
		entries = []IndustryLedgerEntry{}
	}

	statsArgs := []interface{}{userID}
	statsQuery := `
		SELECT
			COUNT(*) AS total,
			SUM(CASE WHEN status IN ('planned', 'queued') THEN 1 ELSE 0 END) AS planned,
			SUM(CASE WHEN status IN ('active', 'paused') THEN 1 ELSE 0 END) AS active,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) AS completed,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) AS failed,
			SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END) AS cancelled,
			COALESCE(SUM(cost_isk), 0) AS total_cost_isk
		  FROM industry_jobs
		 WHERE user_id = ?
	`
	if opt.ProjectID > 0 {
		statsQuery += " AND project_id = ?"
		statsArgs = append(statsArgs, opt.ProjectID)
	}
	if status != "" {
		statsQuery += " AND status = ?"
		statsArgs = append(statsArgs, status)
	}

	var (
		total        sql.NullInt64
		planned      sql.NullInt64
		active       sql.NullInt64
		completed    sql.NullInt64
		failed       sql.NullInt64
		cancelled    sql.NullInt64
		totalCostISK sql.NullFloat64
	)
	if err := d.sql.QueryRow(statsQuery, statsArgs...).Scan(
		&total,
		&planned,
		&active,
		&completed,
		&failed,
		&cancelled,
		&totalCostISK,
	); err != nil {
		return IndustryLedger{}, err
	}

	return IndustryLedger{
		ProjectID:    opt.ProjectID,
		StatusFilter: status,
		Limit:        limit,
		Total:        total.Int64,
		Planned:      planned.Int64,
		Active:       active.Int64,
		Completed:    completed.Int64,
		Failed:       failed.Int64,
		Cancelled:    cancelled.Int64,
		TotalCostISK: totalCostISK.Float64,
		Entries:      entries,
	}, nil
}

func (d *DB) GetIndustryProjectSnapshotForUser(userID string, projectID int64) (IndustryProjectSnapshot, error) {
	userID = normalizeUserID(userID)
	if projectID <= 0 {
		return IndustryProjectSnapshot{}, fmt.Errorf("project_id must be positive")
	}

	project, err := d.GetIndustryProjectForUser(userID, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return IndustryProjectSnapshot{}, sql.ErrNoRows
		}
		return IndustryProjectSnapshot{}, err
	}

	snapshot := IndustryProjectSnapshot{
		Project:      *project,
		Tasks:        []IndustryTask{},
		Jobs:         []IndustryJob{},
		Materials:    []IndustryMaterialPlan{},
		Blueprints:   []IndustryBlueprintPool{},
		MaterialDiff: []IndustryMaterialDiff{},
	}

	taskRows, err := d.sql.Query(`
		SELECT id, user_id, project_id, COALESCE(parent_task_id, 0), name, activity, product_type_id, target_runs,
		       planned_start, planned_end, priority, status, constraints_json, created_at, updated_at
		  FROM industry_tasks
		 WHERE user_id = ? AND project_id = ?
		 ORDER BY priority DESC, id ASC
	`, userID, projectID)
	if err != nil {
		return IndustryProjectSnapshot{}, err
	}
	defer taskRows.Close()

	for taskRows.Next() {
		var (
			row            IndustryTask
			constraintsRaw string
		)
		if err := taskRows.Scan(
			&row.ID,
			&row.UserID,
			&row.ProjectID,
			&row.ParentTaskID,
			&row.Name,
			&row.Activity,
			&row.ProductTypeID,
			&row.TargetRuns,
			&row.PlannedStart,
			&row.PlannedEnd,
			&row.Priority,
			&row.Status,
			&constraintsRaw,
			&row.CreatedAt,
			&row.UpdatedAt,
		); err != nil {
			return IndustryProjectSnapshot{}, err
		}
		row.Constraints = json.RawMessage(normalizeJSONRaw(json.RawMessage(constraintsRaw), "{}"))
		snapshot.Tasks = append(snapshot.Tasks, row)
	}
	if err := taskRows.Err(); err != nil {
		return IndustryProjectSnapshot{}, err
	}

	jobRows, err := d.sql.Query(`
		SELECT id, user_id, project_id, COALESCE(task_id, 0), character_id, facility_id, activity, runs,
		       duration_seconds, cost_isk, status, started_at, finished_at, external_job_id, notes, created_at, updated_at
		  FROM industry_jobs
		 WHERE user_id = ? AND project_id = ?
		 ORDER BY updated_at DESC, id DESC
	`, userID, projectID)
	if err != nil {
		return IndustryProjectSnapshot{}, err
	}
	defer jobRows.Close()

	for jobRows.Next() {
		var row IndustryJob
		if err := jobRows.Scan(
			&row.ID,
			&row.UserID,
			&row.ProjectID,
			&row.TaskID,
			&row.CharacterID,
			&row.FacilityID,
			&row.Activity,
			&row.Runs,
			&row.DurationSeconds,
			&row.CostISK,
			&row.Status,
			&row.StartedAt,
			&row.FinishedAt,
			&row.ExternalJobID,
			&row.Notes,
			&row.CreatedAt,
			&row.UpdatedAt,
		); err != nil {
			return IndustryProjectSnapshot{}, err
		}
		snapshot.Jobs = append(snapshot.Jobs, row)
	}
	if err := jobRows.Err(); err != nil {
		return IndustryProjectSnapshot{}, err
	}

	materialRows, err := d.sql.Query(`
		SELECT id, user_id, project_id, task_id, type_id, type_name, required_qty, available_qty,
		       buy_qty, build_qty, unit_cost_isk, source, updated_at
		  FROM industry_material_plan
		 WHERE user_id = ? AND project_id = ?
		 ORDER BY type_name ASC, type_id ASC, task_id ASC
	`, userID, projectID)
	if err != nil {
		return IndustryProjectSnapshot{}, err
	}
	defer materialRows.Close()

	type materialAgg struct {
		diff IndustryMaterialDiff
	}
	agg := map[int32]*materialAgg{}

	for materialRows.Next() {
		var row IndustryMaterialPlan
		if err := materialRows.Scan(
			&row.ID,
			&row.UserID,
			&row.ProjectID,
			&row.TaskID,
			&row.TypeID,
			&row.TypeName,
			&row.RequiredQty,
			&row.AvailableQty,
			&row.BuyQty,
			&row.BuildQty,
			&row.UnitCostISK,
			&row.Source,
			&row.UpdatedAt,
		); err != nil {
			return IndustryProjectSnapshot{}, err
		}
		snapshot.Materials = append(snapshot.Materials, row)

		item := agg[row.TypeID]
		if item == nil {
			item = &materialAgg{
				diff: IndustryMaterialDiff{
					TypeID:   row.TypeID,
					TypeName: strings.TrimSpace(row.TypeName),
				},
			}
			agg[row.TypeID] = item
		}
		if item.diff.TypeName == "" {
			item.diff.TypeName = strings.TrimSpace(row.TypeName)
		}
		item.diff.RequiredQty += row.RequiredQty
		item.diff.AvailableQty += row.AvailableQty
		item.diff.BuyQty += row.BuyQty
		item.diff.BuildQty += row.BuildQty
	}
	if err := materialRows.Err(); err != nil {
		return IndustryProjectSnapshot{}, err
	}

	blueprintRows, err := d.sql.Query(`
		SELECT id, user_id, project_id, blueprint_type_id, blueprint_name, location_id,
		       quantity, me, te, is_bpo, available_runs, updated_at
		  FROM industry_blueprint_pool
		 WHERE user_id = ? AND project_id = ?
		 ORDER BY blueprint_name ASC, blueprint_type_id ASC, location_id ASC
	`, userID, projectID)
	if err != nil {
		return IndustryProjectSnapshot{}, err
	}
	defer blueprintRows.Close()

	for blueprintRows.Next() {
		var (
			row   IndustryBlueprintPool
			isBPO int
		)
		if err := blueprintRows.Scan(
			&row.ID,
			&row.UserID,
			&row.ProjectID,
			&row.BlueprintTypeID,
			&row.BlueprintName,
			&row.LocationID,
			&row.Quantity,
			&row.ME,
			&row.TE,
			&isBPO,
			&row.AvailableRuns,
			&row.UpdatedAt,
		); err != nil {
			return IndustryProjectSnapshot{}, err
		}
		row.IsBPO = isBPO > 0
		snapshot.Blueprints = append(snapshot.Blueprints, row)
	}
	if err := blueprintRows.Err(); err != nil {
		return IndustryProjectSnapshot{}, err
	}

	diffs := make([]IndustryMaterialDiff, 0, len(agg))
	for _, item := range agg {
		diff := item.diff
		covered := diff.AvailableQty + diff.BuyQty + diff.BuildQty
		missing := diff.RequiredQty - covered
		if missing < 0 {
			missing = 0
		}
		diff.MissingQty = missing
		diffs = append(diffs, diff)
	}
	sort.SliceStable(diffs, func(i, j int) bool {
		if diffs[i].MissingQty != diffs[j].MissingQty {
			return diffs[i].MissingQty > diffs[j].MissingQty
		}
		if diffs[i].RequiredQty != diffs[j].RequiredQty {
			return diffs[i].RequiredQty > diffs[j].RequiredQty
		}
		return diffs[i].TypeID < diffs[j].TypeID
	})
	snapshot.MaterialDiff = diffs

	return snapshot, nil
}
