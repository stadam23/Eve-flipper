package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"eve-flipper/internal/auth"
	"eve-flipper/internal/db"
	"eve-flipper/internal/engine"
	"eve-flipper/internal/sde"
)

var apiTestDBMu sync.Mutex

func openAPITestDB(t *testing.T) *db.DB {
	t.Helper()

	apiTestDBMu.Lock()
	tmpDir := t.TempDir()

	prevWD, err := os.Getwd()
	if err != nil {
		apiTestDBMu.Unlock()
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		apiTestDBMu.Unlock()
		t.Fatalf("chdir temp dir: %v", err)
	}

	database, err := db.Open()
	if err != nil {
		_ = os.Chdir(prevWD)
		apiTestDBMu.Unlock()
		t.Fatalf("open db: %v", err)
	}

	t.Cleanup(func() {
		_ = database.Close()
		_ = os.Chdir(prevWD)
		apiTestDBMu.Unlock()
	})
	return database
}

func requestWithUserID(method, target string, body io.Reader, userID string) *http.Request {
	req := httptest.NewRequest(method, target, body)
	ctx := context.WithValue(req.Context(), userIDContextKey, userID)
	return req.WithContext(ctx)
}

func newAuthedIndustryTestServer(t *testing.T, database *db.DB, userID string) *Server {
	t.Helper()

	sessions := auth.NewSessionStore(database.SqlDB())
	if err := sessions.SaveAndActivateForUser(userID, &auth.Session{
		CharacterID:   90000001,
		CharacterName: "Test Pilot",
		AccessToken:   "test-access-token",
		RefreshToken:  "test-refresh-token",
		ExpiresAt:     time.Now().Add(2 * time.Hour),
	}); err != nil {
		t.Fatalf("SaveAndActivateForUser: %v", err)
	}

	return &Server{
		db:       database,
		sessions: sessions,
	}
}

func TestHandleAuthIndustryProjectSnapshot_Success(t *testing.T) {
	database := openAPITestDB(t)

	userID := "user-api-snapshot"
	srv := newAuthedIndustryTestServer(t, database, userID)
	project, err := database.CreateIndustryProjectForUser(userID, db.IndustryProjectCreateInput{
		Name: "Snapshot API Project",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	_, err = database.ApplyIndustryPlanForUser(userID, project.ID, db.IndustryPlanPatch{
		Replace: true,
		Tasks: []db.IndustryTaskPlanInput{
			{
				Name:       "Build final item",
				Activity:   "manufacturing",
				TargetRuns: 3,
			},
		},
		Jobs: []db.IndustryJobPlanInput{
			{
				Activity:        "manufacturing",
				Runs:            3,
				DurationSeconds: 1800,
				CostISK:         500000,
				Status:          db.IndustryJobStatusPlanned,
				Notes:           "seeded for snapshot test",
			},
		},
		Materials: []db.IndustryMaterialPlanInput{
			{
				TypeID:       34,
				TypeName:     "Tritanium",
				RequiredQty:  100,
				AvailableQty: 10,
				BuyQty:       60,
				BuildQty:     10,
				UnitCostISK:  5.2,
				Source:       "market",
			},
		},
		Blueprints: []db.IndustryBlueprintPoolInput{
			{
				BlueprintTypeID: 700001,
				BlueprintName:   "Example Blueprint",
				LocationID:      60003760,
				Quantity:        1,
				ME:              10,
				TE:              20,
				IsBPO:           true,
				AvailableRuns:   0,
			},
		},
	})
	if err != nil {
		t.Fatalf("ApplyIndustryPlanForUser: %v", err)
	}

	req := requestWithUserID(http.MethodGet, "/api/auth/industry/projects/"+strconv.FormatInt(project.ID, 10)+"/snapshot", nil, userID)
	req.SetPathValue("projectID", strconv.FormatInt(project.ID, 10))
	rec := httptest.NewRecorder()

	srv.handleAuthIndustryProjectSnapshot(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var out db.IndustryProjectSnapshot
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}

	if out.Project.ID != project.ID {
		t.Fatalf("project id = %d, want %d", out.Project.ID, project.ID)
	}
	if len(out.Tasks) != 1 {
		t.Fatalf("tasks len = %d, want 1", len(out.Tasks))
	}
	if len(out.Jobs) != 1 {
		t.Fatalf("jobs len = %d, want 1", len(out.Jobs))
	}
	if len(out.Materials) != 1 {
		t.Fatalf("materials len = %d, want 1", len(out.Materials))
	}
	if len(out.Blueprints) != 1 {
		t.Fatalf("blueprints len = %d, want 1", len(out.Blueprints))
	}
	if len(out.MaterialDiff) != 1 {
		t.Fatalf("material_diff len = %d, want 1", len(out.MaterialDiff))
	}
	if out.MaterialDiff[0].MissingQty != 20 {
		t.Fatalf("missing_qty = %d, want 20", out.MaterialDiff[0].MissingQty)
	}
}

func TestHandleAuthBulkUpdateIndustryJobStatus_Success(t *testing.T) {
	database := openAPITestDB(t)

	userID := "user-api-bulk"
	srv := newAuthedIndustryTestServer(t, database, userID)
	project, err := database.CreateIndustryProjectForUser(userID, db.IndustryProjectCreateInput{
		Name: "Bulk API Project",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	_, err = database.ApplyIndustryPlanForUser(userID, project.ID, db.IndustryPlanPatch{
		Replace: true,
		Jobs: []db.IndustryJobPlanInput{
			{Activity: "manufacturing", Runs: 1, DurationSeconds: 1200, CostISK: 100000, Status: db.IndustryJobStatusPlanned},
			{Activity: "manufacturing", Runs: 2, DurationSeconds: 1800, CostISK: 120000, Status: db.IndustryJobStatusPlanned},
		},
	})
	if err != nil {
		t.Fatalf("ApplyIndustryPlanForUser: %v", err)
	}

	ledgerBefore, err := database.GetIndustryLedgerForUser(userID, db.IndustryLedgerOptions{
		ProjectID: project.ID,
		Limit:     20,
	})
	if err != nil {
		t.Fatalf("GetIndustryLedgerForUser: %v", err)
	}
	if len(ledgerBefore.Entries) != 2 {
		t.Fatalf("ledger entries len = %d, want 2", len(ledgerBefore.Entries))
	}
	jobIDs := []int64{ledgerBefore.Entries[0].JobID, ledgerBefore.Entries[1].JobID}

	payload := map[string]interface{}{
		"job_ids": jobIDs,
		"status":  db.IndustryJobStatusCompleted,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := requestWithUserID(http.MethodPatch, "/api/auth/industry/jobs/status/bulk", bytes.NewReader(body), userID)
	rec := httptest.NewRecorder()

	srv.handleAuthBulkUpdateIndustryJobStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var out struct {
		OK      bool             `json:"ok"`
		Updated int              `json:"updated"`
		Jobs    []db.IndustryJob `json:"jobs"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode bulk response: %v", err)
	}
	if !out.OK {
		t.Fatalf("ok = false, want true")
	}
	if out.Updated != 2 {
		t.Fatalf("updated = %d, want 2", out.Updated)
	}
	if len(out.Jobs) != 2 {
		t.Fatalf("jobs len = %d, want 2", len(out.Jobs))
	}
	for _, job := range out.Jobs {
		if job.Status != db.IndustryJobStatusCompleted {
			t.Fatalf("job status = %q, want completed", job.Status)
		}
	}

	ledgerAfter, err := database.GetIndustryLedgerForUser(userID, db.IndustryLedgerOptions{
		ProjectID: project.ID,
		Limit:     20,
	})
	if err != nil {
		t.Fatalf("GetIndustryLedgerForUser after: %v", err)
	}
	if ledgerAfter.Completed != 2 {
		t.Fatalf("ledger completed = %d, want 2", ledgerAfter.Completed)
	}
}

func TestHandleAuthBulkUpdateIndustryJobStatus_Validation(t *testing.T) {
	database := openAPITestDB(t)
	srv := newAuthedIndustryTestServer(t, database, "user-api-invalid")

	req := requestWithUserID(http.MethodPatch, "/api/auth/industry/jobs/status/bulk", bytes.NewBufferString(`{"job_ids":[],"status":"completed"}`), "user-api-invalid")
	rec := httptest.NewRecorder()

	srv.handleAuthBulkUpdateIndustryJobStatus(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandleAuthIndustryEndpoints_RequireLogin(t *testing.T) {
	database := openAPITestDB(t)
	srv := &Server{db: database}

	req := requestWithUserID(http.MethodGet, "/api/auth/industry/projects", nil, "user-no-session")
	rec := httptest.NewRecorder()

	srv.handleAuthListIndustryProjects(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandleAuthPreviewIndustryProjectPlan_Success(t *testing.T) {
	database := openAPITestDB(t)
	userID := "user-api-preview-plan"
	srv := newAuthedIndustryTestServer(t, database, userID)

	project, err := database.CreateIndustryProjectForUser(userID, db.IndustryProjectCreateInput{
		Name: "Preview Plan Project",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	payload := db.IndustryPlanPatch{
		Replace: true,
		Tasks: []db.IndustryTaskPlanInput{
			{Name: "Preview Task", Activity: "manufacturing", TargetRuns: 2},
		},
		Jobs: []db.IndustryJobPlanInput{
			{
				TaskID:          1,
				Activity:        "manufacturing",
				Runs:            2,
				DurationSeconds: 1800,
				CostISK:         150000,
				Status:          db.IndustryJobStatusPlanned,
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	req := requestWithUserID(http.MethodPost, "/api/auth/industry/projects/"+strconv.FormatInt(project.ID, 10)+"/plan/preview", bytes.NewReader(body), userID)
	req.SetPathValue("projectID", strconv.FormatInt(project.ID, 10))
	rec := httptest.NewRecorder()

	srv.handleAuthPreviewIndustryProjectPlan(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var out db.IndustryPlanPreview
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode preview response: %v", err)
	}
	if out.ProjectID != project.ID {
		t.Fatalf("project_id = %d, want %d", out.ProjectID, project.ID)
	}
	if out.Summary.TasksInserted != 1 {
		t.Fatalf("tasks_inserted = %d, want 1", out.Summary.TasksInserted)
	}
	if len(out.Tasks) != 1 {
		t.Fatalf("preview tasks len = %d, want 1", len(out.Tasks))
	}
}

func TestHandleAuthPlanIndustryProject_Success(t *testing.T) {
	database := openAPITestDB(t)
	userID := "user-api-apply-plan"
	srv := newAuthedIndustryTestServer(t, database, userID)

	project, err := database.CreateIndustryProjectForUser(userID, db.IndustryProjectCreateInput{
		Name: "Apply Plan Project",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	payload := db.IndustryPlanPatch{
		Replace: true,
		Tasks: []db.IndustryTaskPlanInput{
			{Name: "Applied Task", Activity: "manufacturing", TargetRuns: 1},
		},
		Jobs: []db.IndustryJobPlanInput{
			{
				TaskID:          1,
				Activity:        "manufacturing",
				Runs:            1,
				DurationSeconds: 900,
				CostISK:         75000,
				Status:          db.IndustryJobStatusPlanned,
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	req := requestWithUserID(http.MethodPost, "/api/auth/industry/projects/"+strconv.FormatInt(project.ID, 10)+"/plan", bytes.NewReader(body), userID)
	req.SetPathValue("projectID", strconv.FormatInt(project.ID, 10))
	rec := httptest.NewRecorder()

	srv.handleAuthPlanIndustryProject(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		OK      bool                   `json:"ok"`
		Summary db.IndustryPlanSummary `json:"summary"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode apply response: %v", err)
	}
	if !out.OK {
		t.Fatalf("ok = false, want true")
	}
	if out.Summary.TasksInserted != 1 {
		t.Fatalf("tasks_inserted = %d, want 1", out.Summary.TasksInserted)
	}
	if out.Summary.JobsInserted != 1 {
		t.Fatalf("jobs_inserted = %d, want 1", out.Summary.JobsInserted)
	}
}

func TestHandleAuthIndustryLedger_Success(t *testing.T) {
	database := openAPITestDB(t)
	userID := "user-api-ledger"
	srv := newAuthedIndustryTestServer(t, database, userID)

	project, err := database.CreateIndustryProjectForUser(userID, db.IndustryProjectCreateInput{
		Name: "Ledger Project",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}
	if _, err := database.ApplyIndustryPlanForUser(userID, project.ID, db.IndustryPlanPatch{
		Replace: true,
		Jobs: []db.IndustryJobPlanInput{
			{
				Activity:        "manufacturing",
				Runs:            1,
				DurationSeconds: 600,
				CostISK:         42000,
				Status:          db.IndustryJobStatusPlanned,
			},
		},
	}); err != nil {
		t.Fatalf("ApplyIndustryPlanForUser: %v", err)
	}

	req := requestWithUserID(http.MethodGet, "/api/auth/industry/ledger?project_id="+strconv.FormatInt(project.ID, 10)+"&limit=20", nil, userID)
	rec := httptest.NewRecorder()

	srv.handleAuthIndustryLedger(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var out db.IndustryLedger
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode ledger response: %v", err)
	}
	if out.ProjectID != project.ID {
		t.Fatalf("ledger project_id = %d, want %d", out.ProjectID, project.ID)
	}
	if len(out.Entries) != 1 {
		t.Fatalf("ledger entries len = %d, want 1", len(out.Entries))
	}
}

func TestCORSMiddleware_AllowsPATCH(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/auth/industry/jobs/status", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "PATCH")
	req.Host = "localhost:8080"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	allowMethods := rec.Header().Get("Access-Control-Allow-Methods")
	if !strings.Contains(allowMethods, "PATCH") {
		t.Fatalf("Access-Control-Allow-Methods = %q, want to contain PATCH", allowMethods)
	}
}

func TestHandleIndustrySearch_LimitIsClamped(t *testing.T) {
	analyzer := &engine.IndustryAnalyzer{
		SDE: &sde.Data{
			Types: map[int32]*sde.ItemType{},
		},
	}
	for i := 1; i <= 250; i++ {
		typeID := int32(i)
		analyzer.SDE.Types[typeID] = &sde.ItemType{
			ID:   typeID,
			Name: "Item " + strconv.Itoa(i),
		}
	}

	srv := &Server{
		ready:            true,
		industryAnalyzer: analyzer,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/industry/search?q=item&limit=9999", nil)
	rec := httptest.NewRecorder()
	srv.handleIndustrySearch(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var out []map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode search response: %v", err)
	}
	if len(out) != industrySearchMaxLimit {
		t.Fatalf("search result len = %d, want %d (clamped)", len(out), industrySearchMaxLimit)
	}
}

func TestHandleIndustryAnalyze_RejectsOversizedBody(t *testing.T) {
	srv := &Server{ready: true}
	oversized := strings.Repeat("x", industryAnalyzeMaxBodyBytes+32)
	body := `{"type_id":1,"system_name":"` + oversized + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/industry/analyze", strings.NewReader(body))
	rec := httptest.NewRecorder()

	srv.handleIndustryAnalyze(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandleIndustryAnalyze_RejectsNegativeTypeID(t *testing.T) {
	srv := &Server{ready: true}
	req := httptest.NewRequest(http.MethodPost, "/api/industry/analyze", strings.NewReader(`{"type_id":-7}`))
	rec := httptest.NewRecorder()

	srv.handleIndustryAnalyze(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}
