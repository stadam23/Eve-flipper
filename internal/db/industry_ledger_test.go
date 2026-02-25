package db

import (
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestIndustryLedgerProjectCreateAndListByUser(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	_, err := d.CreateIndustryProjectForUser("user-a", IndustryProjectCreateInput{
		Name:     "Capital Components",
		Strategy: "balanced",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser user-a: %v", err)
	}
	_, err = d.CreateIndustryProjectForUser("user-b", IndustryProjectCreateInput{
		Name:     "Reactions Batch",
		Strategy: "aggressive",
		Status:   IndustryProjectStatusActive,
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser user-b: %v", err)
	}

	userAProjects, err := d.ListIndustryProjectsForUser("user-a", "", 50)
	if err != nil {
		t.Fatalf("ListIndustryProjectsForUser user-a: %v", err)
	}
	if len(userAProjects) != 1 {
		t.Fatalf("user-a projects len = %d, want 1", len(userAProjects))
	}
	if userAProjects[0].Name != "Capital Components" {
		t.Fatalf("user-a project name = %q, want Capital Components", userAProjects[0].Name)
	}

	activeForUserA, err := d.ListIndustryProjectsForUser("user-a", IndustryProjectStatusActive, 50)
	if err != nil {
		t.Fatalf("ListIndustryProjectsForUser active user-a: %v", err)
	}
	if len(activeForUserA) != 0 {
		t.Fatalf("user-a active projects len = %d, want 0", len(activeForUserA))
	}
}

func TestIndustryLedgerApplyPlanAndReadLedger(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	project, err := d.CreateIndustryProjectForUser("user-x", IndustryProjectCreateInput{
		Name:     "T2 Module Pipeline",
		Strategy: "conservative",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	plan := IndustryPlanPatch{
		Replace:       true,
		ProjectStatus: IndustryProjectStatusPlanned,
		Tasks: []IndustryTaskPlanInput{
			{
				Name:          "Build input reactions",
				Activity:      "reaction",
				ProductTypeID: 12345,
				TargetRuns:    20,
				Priority:      90,
				Status:        IndustryTaskStatusReady,
			},
			{
				Name:          "Manufacture final module",
				Activity:      "manufacturing",
				ProductTypeID: 54321,
				TargetRuns:    10,
				Priority:      100,
				Status:        IndustryTaskStatusPlanned,
			},
		},
		Jobs: []IndustryJobPlanInput{
			{
				Activity:        "reaction",
				Runs:            20,
				DurationSeconds: 7200,
				CostISK:         4_200_000,
				Status:          IndustryJobStatusPlanned,
			},
			{
				Activity:        "manufacturing",
				Runs:            10,
				DurationSeconds: 14400,
				CostISK:         9_100_000,
				Status:          IndustryJobStatusQueued,
			},
		},
		Materials: []IndustryMaterialPlanInput{
			{
				TypeID:       34,
				TypeName:     "Tritanium",
				RequiredQty:  400000,
				AvailableQty: 150000,
				BuyQty:       250000,
				BuildQty:     0,
				UnitCostISK:  5.2,
				Source:       "market",
			},
			{
				TypeID:       35,
				TypeName:     "Pyerite",
				RequiredQty:  180000,
				AvailableQty: 90000,
				BuyQty:       90000,
				BuildQty:     0,
				UnitCostISK:  10.1,
				Source:       "stock",
			},
		},
		Blueprints: []IndustryBlueprintPoolInput{
			{
				BlueprintTypeID: 777001,
				BlueprintName:   "Example Module Blueprint",
				LocationID:      60003760,
				Quantity:        1,
				ME:              10,
				TE:              20,
				IsBPO:           true,
				AvailableRuns:   0,
			},
		},
	}
	summary, err := d.ApplyIndustryPlanForUser("user-x", project.ID, plan)
	if err != nil {
		t.Fatalf("ApplyIndustryPlanForUser: %v", err)
	}
	if summary.TasksInserted != 2 || summary.JobsInserted != 2 || summary.MaterialsUpsert != 2 || summary.BlueprintsUpsert != 1 {
		t.Fatalf("unexpected summary counts: %+v", summary)
	}
	if summary.ProjectStatus != IndustryProjectStatusPlanned {
		t.Fatalf("project status = %q, want planned", summary.ProjectStatus)
	}

	ledger, err := d.GetIndustryLedgerForUser("user-x", IndustryLedgerOptions{
		ProjectID: project.ID,
		Limit:     50,
	})
	if err != nil {
		t.Fatalf("GetIndustryLedgerForUser: %v", err)
	}
	if ledger.Total != 2 {
		t.Fatalf("ledger total = %d, want 2", ledger.Total)
	}
	if len(ledger.Entries) != 2 {
		t.Fatalf("ledger entries len = %d, want 2", len(ledger.Entries))
	}
	if ledger.TotalCostISK <= 0 {
		t.Fatalf("ledger total cost = %f, want > 0", ledger.TotalCostISK)
	}
}

func TestIndustryLedgerUpdateJobStatus(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	project, err := d.CreateIndustryProjectForUser("user-z", IndustryProjectCreateInput{
		Name: "Status Test Project",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	_, err = d.ApplyIndustryPlanForUser("user-z", project.ID, IndustryPlanPatch{
		Replace: true,
		Jobs: []IndustryJobPlanInput{
			{
				Activity:        "manufacturing",
				Runs:            5,
				DurationSeconds: 3600,
				CostISK:         1_000_000,
				Status:          IndustryJobStatusPlanned,
			},
		},
	})
	if err != nil {
		t.Fatalf("ApplyIndustryPlanForUser: %v", err)
	}

	ledger, err := d.GetIndustryLedgerForUser("user-z", IndustryLedgerOptions{ProjectID: project.ID, Limit: 10})
	if err != nil {
		t.Fatalf("GetIndustryLedgerForUser: %v", err)
	}
	if len(ledger.Entries) != 1 {
		t.Fatalf("ledger entries len = %d, want 1", len(ledger.Entries))
	}
	jobID := ledger.Entries[0].JobID

	activeJob, err := d.UpdateIndustryJobStatusForUser("user-z", jobID, IndustryJobStatusActive, "", "", "running")
	if err != nil {
		t.Fatalf("UpdateIndustryJobStatusForUser active: %v", err)
	}
	if activeJob.Status != IndustryJobStatusActive {
		t.Fatalf("job status after active = %q, want active", activeJob.Status)
	}
	if activeJob.StartedAt == "" {
		t.Fatalf("job started_at should be auto-filled for active status")
	}

	completedJob, err := d.UpdateIndustryJobStatusForUser("user-z", jobID, IndustryJobStatusCompleted, "", "", "")
	if err != nil {
		t.Fatalf("UpdateIndustryJobStatusForUser completed: %v", err)
	}
	if completedJob.Status != IndustryJobStatusCompleted {
		t.Fatalf("job status after completed = %q, want completed", completedJob.Status)
	}
	if completedJob.FinishedAt == "" {
		t.Fatalf("job finished_at should be auto-filled for completed status")
	}

	_, err = d.UpdateIndustryJobStatusForUser("user-other", jobID, IndustryJobStatusCancelled, "", "", "")
	if err == nil {
		t.Fatalf("expected error for cross-user job update")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("cross-user update err = %v, want sql.ErrNoRows", err)
	}
}

func TestIndustryLedgerBulkUpdateJobStatus(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	project, err := d.CreateIndustryProjectForUser("user-bulk-status", IndustryProjectCreateInput{
		Name: "Bulk Status Project",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	_, err = d.ApplyIndustryPlanForUser("user-bulk-status", project.ID, IndustryPlanPatch{
		Replace: true,
		Jobs: []IndustryJobPlanInput{
			{Activity: "manufacturing", Runs: 2, DurationSeconds: 600, CostISK: 1000, Status: IndustryJobStatusPlanned},
			{Activity: "manufacturing", Runs: 3, DurationSeconds: 900, CostISK: 1500, Status: IndustryJobStatusPlanned},
			{Activity: "manufacturing", Runs: 4, DurationSeconds: 1200, CostISK: 2000, Status: IndustryJobStatusPlanned},
		},
	})
	if err != nil {
		t.Fatalf("ApplyIndustryPlanForUser: %v", err)
	}

	ledgerBefore, err := d.GetIndustryLedgerForUser("user-bulk-status", IndustryLedgerOptions{
		ProjectID: project.ID,
		Limit:     20,
	})
	if err != nil {
		t.Fatalf("GetIndustryLedgerForUser before: %v", err)
	}
	if len(ledgerBefore.Entries) != 3 {
		t.Fatalf("len(ledgerBefore.Entries) = %d, want 3", len(ledgerBefore.Entries))
	}

	firstID := ledgerBefore.Entries[0].JobID
	secondID := ledgerBefore.Entries[1].JobID

	updatedJobs, err := d.UpdateIndustryJobStatusesForUser(
		"user-bulk-status",
		[]int64{firstID, secondID, firstID}, // duplicate ID should be deduplicated
		IndustryJobStatusCompleted,
		"",
		"",
		"bulk update",
	)
	if err != nil {
		t.Fatalf("UpdateIndustryJobStatusesForUser: %v", err)
	}
	if len(updatedJobs) != 2 {
		t.Fatalf("len(updatedJobs) = %d, want 2", len(updatedJobs))
	}
	for _, j := range updatedJobs {
		if j.Status != IndustryJobStatusCompleted {
			t.Fatalf("job status = %q, want completed", j.Status)
		}
		if j.FinishedAt == "" {
			t.Fatalf("finished_at should be auto-filled for completed bulk update")
		}
		if j.Notes != "bulk update" {
			t.Fatalf("notes = %q, want bulk update", j.Notes)
		}
	}

	ledgerAfter, err := d.GetIndustryLedgerForUser("user-bulk-status", IndustryLedgerOptions{
		ProjectID: project.ID,
		Limit:     20,
	})
	if err != nil {
		t.Fatalf("GetIndustryLedgerForUser after: %v", err)
	}
	if ledgerAfter.Completed != 2 {
		t.Fatalf("ledgerAfter.Completed = %d, want 2", ledgerAfter.Completed)
	}
	if ledgerAfter.Planned != 1 {
		t.Fatalf("ledgerAfter.Planned = %d, want 1", ledgerAfter.Planned)
	}

	_, err = d.UpdateIndustryJobStatusesForUser(
		"user-bulk-status",
		[]int64{firstID, 999999999},
		IndustryJobStatusCancelled,
		"",
		"",
		"",
	)
	if err == nil {
		t.Fatalf("expected error when one or more job IDs do not exist")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("err = %v, want sql.ErrNoRows", err)
	}
}

func TestIndustryLedgerSchedulerSplitAndQueue(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	project, err := d.CreateIndustryProjectForUser("user-sched", IndustryProjectCreateInput{
		Name: "Scheduler Test",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	summary, err := d.ApplyIndustryPlanForUser("user-sched", project.ID, IndustryPlanPatch{
		Replace: true,
		Jobs: []IndustryJobPlanInput{
			{
				Activity:        "manufacturing",
				Runs:            25,
				DurationSeconds: 25000,
				CostISK:         1_250_000,
				Status:          IndustryJobStatusPlanned,
			},
		},
		Scheduler: IndustryPlanSchedulerInput{
			Enabled:               true,
			SlotCount:             2,
			MaxJobRuns:            6,
			MaxJobDurationSeconds: 7200,
			WindowDays:            30,
			QueueStatus:           IndustryJobStatusQueued,
		},
	})
	if err != nil {
		t.Fatalf("ApplyIndustryPlanForUser with scheduler: %v", err)
	}
	if !summary.SchedulerApplied {
		t.Fatalf("summary.SchedulerApplied = false, want true")
	}
	if summary.JobsSplitFrom != 1 {
		t.Fatalf("summary.JobsSplitFrom = %d, want 1", summary.JobsSplitFrom)
	}
	if summary.JobsPlannedTotal != 5 {
		t.Fatalf("summary.JobsPlannedTotal = %d, want 5", summary.JobsPlannedTotal)
	}
	if summary.JobsInserted != 5 {
		t.Fatalf("summary.JobsInserted = %d, want 5", summary.JobsInserted)
	}

	ledger, err := d.GetIndustryLedgerForUser("user-sched", IndustryLedgerOptions{
		ProjectID: project.ID,
		Limit:     50,
	})
	if err != nil {
		t.Fatalf("GetIndustryLedgerForUser: %v", err)
	}
	if ledger.Total != 5 {
		t.Fatalf("ledger.Total = %d, want 5", ledger.Total)
	}
	if len(ledger.Entries) != 5 {
		t.Fatalf("len(ledger.Entries) = %d, want 5", len(ledger.Entries))
	}
	for _, e := range ledger.Entries {
		if e.Status != IndustryJobStatusQueued {
			t.Fatalf("job status = %q, want queued", e.Status)
		}
		if e.StartedAt == "" {
			t.Fatalf("scheduled job started_at is empty")
		}
		if e.Runs <= 0 || e.Runs > 6 {
			t.Fatalf("scheduled job runs = %d, want 1..6", e.Runs)
		}
	}
}

func TestIndustryLedgerSchedulerRespectsTaskDependencies(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	project, err := d.CreateIndustryProjectForUser("user-deps", IndustryProjectCreateInput{
		Name: "Dependency Scheduler Test",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	summary, err := d.ApplyIndustryPlanForUser("user-deps", project.ID, IndustryPlanPatch{
		Replace: true,
		Tasks: []IndustryTaskPlanInput{
			{
				Name:       "Parent Task",
				Activity:   "manufacturing",
				TargetRuns: 2,
				Constraints: json.RawMessage(`{
					"duration_seconds_per_run": 600,
					"cost_isk_per_run": 1000
				}`),
			},
			{
				Name:         "Child Task",
				ParentTaskID: 1,
				Activity:     "manufacturing",
				TargetRuns:   1,
				Constraints: json.RawMessage(`{
					"duration_seconds_per_run": 300,
					"cost_isk_per_run": 500
				}`),
			},
		},
		Scheduler: IndustryPlanSchedulerInput{
			Enabled:               true,
			SlotCount:             2,
			MaxJobRuns:            10,
			MaxJobDurationSeconds: 7200,
			QueueStatus:           IndustryJobStatusQueued,
		},
	})
	if err != nil {
		t.Fatalf("ApplyIndustryPlanForUser scheduler deps: %v", err)
	}
	if !summary.SchedulerApplied {
		t.Fatalf("summary.SchedulerApplied = false, want true")
	}

	ledger, err := d.GetIndustryLedgerForUser("user-deps", IndustryLedgerOptions{
		ProjectID: project.ID,
		Limit:     20,
	})
	if err != nil {
		t.Fatalf("GetIndustryLedgerForUser: %v", err)
	}
	if len(ledger.Entries) != 2 {
		t.Fatalf("len(ledger.Entries) = %d, want 2", len(ledger.Entries))
	}

	var parentEntry *IndustryLedgerEntry
	var childEntry *IndustryLedgerEntry
	for i := range ledger.Entries {
		entry := &ledger.Entries[i]
		switch entry.TaskName {
		case "Parent Task":
			parentEntry = entry
		case "Child Task":
			childEntry = entry
		}
	}
	if parentEntry == nil || childEntry == nil {
		t.Fatalf("failed to locate parent/child entries in ledger: %+v", ledger.Entries)
	}
	parentFinish, err := time.Parse(time.RFC3339, parentEntry.FinishedAt)
	if err != nil {
		t.Fatalf("parent finished_at parse: %v", err)
	}
	childStart, err := time.Parse(time.RFC3339, childEntry.StartedAt)
	if err != nil {
		t.Fatalf("child started_at parse: %v", err)
	}
	if childStart.Before(parentFinish) {
		t.Fatalf("dependency violation: child_start=%s before parent_finish=%s", childStart.Format(time.RFC3339), parentFinish.Format(time.RFC3339))
	}
}

func TestIndustryLedgerPreviewPlanNoPersist(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	project, err := d.CreateIndustryProjectForUser("user-preview", IndustryProjectCreateInput{
		Name: "Preview Test",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	preview, err := d.PreviewIndustryPlanForUser("user-preview", project.ID, IndustryPlanPatch{
		Replace: true,
		Tasks: []IndustryTaskPlanInput{
			{
				Name:       "Preview Task",
				Activity:   "manufacturing",
				TargetRuns: 4,
				Constraints: json.RawMessage(`{
					"duration_seconds_per_run": 120,
					"cost_isk_per_run": 100
				}`),
			},
		},
		Scheduler: IndustryPlanSchedulerInput{
			Enabled:               true,
			SlotCount:             1,
			MaxJobRuns:            2,
			MaxJobDurationSeconds: 3600,
			QueueStatus:           IndustryJobStatusQueued,
		},
	})
	if err != nil {
		t.Fatalf("PreviewIndustryPlanForUser: %v", err)
	}
	if preview.Summary.TasksInserted != 1 {
		t.Fatalf("preview tasks_inserted = %d, want 1", preview.Summary.TasksInserted)
	}
	if preview.Summary.JobsInserted != 2 {
		t.Fatalf("preview jobs_inserted = %d, want 2", preview.Summary.JobsInserted)
	}
	if len(preview.Jobs) != 2 {
		t.Fatalf("preview jobs len = %d, want 2", len(preview.Jobs))
	}

	// Preview must not persist anything.
	ledger, err := d.GetIndustryLedgerForUser("user-preview", IndustryLedgerOptions{ProjectID: project.ID, Limit: 10})
	if err != nil {
		t.Fatalf("GetIndustryLedgerForUser: %v", err)
	}
	if ledger.Total != 0 {
		t.Fatalf("ledger total after preview = %d, want 0", ledger.Total)
	}
}

func TestIndustryLedgerSchedulerStrategyDefaults(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	conservative, err := d.CreateIndustryProjectForUser("user-strategy", IndustryProjectCreateInput{
		Name:     "Conservative Scheduler",
		Strategy: "conservative",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser conservative: %v", err)
	}

	summaryConservative, err := d.ApplyIndustryPlanForUser("user-strategy", conservative.ID, IndustryPlanPatch{
		Replace: true,
		Jobs: []IndustryJobPlanInput{
			{
				Activity:        "manufacturing",
				Runs:            500,
				DurationSeconds: 50000,
				CostISK:         10_000_000,
				Status:          IndustryJobStatusPlanned,
			},
		},
		Scheduler: IndustryPlanSchedulerInput{
			Enabled: true, // rely on strategy defaults for slot/runs/duration/status
		},
	})
	if err != nil {
		t.Fatalf("ApplyIndustryPlanForUser conservative scheduler: %v", err)
	}
	if !summaryConservative.SchedulerApplied {
		t.Fatalf("conservative scheduler_applied = false, want true")
	}
	if summaryConservative.JobsInserted != 10 {
		t.Fatalf("conservative jobs_inserted = %d, want 10 (runs split by profile max_job_runs=50)", summaryConservative.JobsInserted)
	}

	ledgerConservative, err := d.GetIndustryLedgerForUser("user-strategy", IndustryLedgerOptions{
		ProjectID: conservative.ID,
		Limit:     50,
	})
	if err != nil {
		t.Fatalf("GetIndustryLedgerForUser conservative: %v", err)
	}
	for _, e := range ledgerConservative.Entries {
		if e.Status != IndustryJobStatusPlanned {
			t.Fatalf("conservative scheduled status = %q, want planned", e.Status)
		}
		if e.Runs <= 0 || e.Runs > 50 {
			t.Fatalf("conservative scheduled runs = %d, want 1..50", e.Runs)
		}
	}

	aggressive, err := d.CreateIndustryProjectForUser("user-strategy", IndustryProjectCreateInput{
		Name:     "Aggressive Scheduler",
		Strategy: "aggressive",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser aggressive: %v", err)
	}

	summaryAggressive, err := d.ApplyIndustryPlanForUser("user-strategy", aggressive.ID, IndustryPlanPatch{
		Replace: true,
		Jobs: []IndustryJobPlanInput{
			{
				Activity:        "manufacturing",
				Runs:            500,
				DurationSeconds: 50000,
				CostISK:         10_000_000,
				Status:          IndustryJobStatusPlanned,
			},
		},
		Scheduler: IndustryPlanSchedulerInput{
			Enabled: true, // rely on strategy defaults
		},
	})
	if err != nil {
		t.Fatalf("ApplyIndustryPlanForUser aggressive scheduler: %v", err)
	}
	if !summaryAggressive.SchedulerApplied {
		t.Fatalf("aggressive scheduler_applied = false, want true")
	}
	if summaryAggressive.JobsInserted != 2 {
		t.Fatalf("aggressive jobs_inserted = %d, want 2 (runs split by profile max_job_runs=400)", summaryAggressive.JobsInserted)
	}

	ledgerAggressive, err := d.GetIndustryLedgerForUser("user-strategy", IndustryLedgerOptions{
		ProjectID: aggressive.ID,
		Limit:     20,
	})
	if err != nil {
		t.Fatalf("GetIndustryLedgerForUser aggressive: %v", err)
	}
	for _, e := range ledgerAggressive.Entries {
		if e.Status != IndustryJobStatusQueued {
			t.Fatalf("aggressive scheduled status = %q, want queued", e.Status)
		}
		if e.Runs <= 0 || e.Runs > 400 {
			t.Fatalf("aggressive scheduled runs = %d, want 1..400", e.Runs)
		}
	}
}

func TestIndustryLedgerSchedulerDerivesMultiActivityJobsFromTasks(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	project, err := d.CreateIndustryProjectForUser("user-multi-activity", IndustryProjectCreateInput{
		Name:     "Multi Activity Plan",
		Strategy: "balanced",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	summary, err := d.ApplyIndustryPlanForUser("user-multi-activity", project.ID, IndustryPlanPatch{
		Replace: true,
		Tasks: []IndustryTaskPlanInput{
			{Name: "Manufacturing chain", Activity: "manufacturing", TargetRuns: 2},
			{Name: "Reaction chain", Activity: "reaction", TargetRuns: 3},
			{Name: "Copy chain", Activity: "copy", TargetRuns: 1},
			{Name: "Invention chain", Activity: "invention", TargetRuns: 1},
		},
		Scheduler: IndustryPlanSchedulerInput{
			Enabled: true,
		},
	})
	if err != nil {
		t.Fatalf("ApplyIndustryPlanForUser multi-activity scheduler: %v", err)
	}
	if !summary.SchedulerApplied {
		t.Fatalf("summary.SchedulerApplied = false, want true")
	}
	if summary.TasksInserted != 4 {
		t.Fatalf("summary.TasksInserted = %d, want 4", summary.TasksInserted)
	}
	if summary.JobsInserted != 4 {
		t.Fatalf("summary.JobsInserted = %d, want 4", summary.JobsInserted)
	}

	ledger, err := d.GetIndustryLedgerForUser("user-multi-activity", IndustryLedgerOptions{
		ProjectID: project.ID,
		Limit:     20,
	})
	if err != nil {
		t.Fatalf("GetIndustryLedgerForUser: %v", err)
	}

	seen := map[string]bool{}
	for _, e := range ledger.Entries {
		seen[e.Activity] = true
		if e.Status != IndustryJobStatusQueued {
			t.Fatalf("derived multi-activity status = %q, want queued", e.Status)
		}
	}
	for _, activity := range []string{"manufacturing", "reaction", "copy", "invention"} {
		if !seen[activity] {
			t.Fatalf("missing derived job for activity %q", activity)
		}
	}
}

func TestIndustryLedgerSchedulerWindowCapsTimeline(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	project, err := d.CreateIndustryProjectForUser("user-window", IndustryProjectCreateInput{
		Name:     "Window Cap Scheduler",
		Strategy: "balanced",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	summary, err := d.ApplyIndustryPlanForUser("user-window", project.ID, IndustryPlanPatch{
		Replace: true,
		Jobs: []IndustryJobPlanInput{
			{
				Activity:        "manufacturing",
				Runs:            10,
				DurationSeconds: 10 * 12 * 3600, // 12h/run
				CostISK:         1_000_000,
				Status:          IndustryJobStatusPlanned,
			},
		},
		Scheduler: IndustryPlanSchedulerInput{
			Enabled:               true,
			SlotCount:             1,
			MaxJobRuns:            10,
			MaxJobDurationSeconds: 24 * 3600, // per-job cap => 2 runs per chunk
			WindowDays:            2,         // total schedule cap
			QueueStatus:           IndustryJobStatusQueued,
		},
	})
	if err != nil {
		t.Fatalf("ApplyIndustryPlanForUser scheduler window: %v", err)
	}
	if !summary.SchedulerApplied {
		t.Fatalf("summary.SchedulerApplied = false, want true")
	}
	if summary.JobsInserted != 3 {
		t.Fatalf("summary.JobsInserted = %d, want 3 (2 scheduled chunks + 1 deferred chunk)", summary.JobsInserted)
	}

	ledger, err := d.GetIndustryLedgerForUser("user-window", IndustryLedgerOptions{
		ProjectID: project.ID,
		Limit:     20,
	})
	if err != nil {
		t.Fatalf("GetIndustryLedgerForUser: %v", err)
	}
	if len(ledger.Entries) != 3 {
		t.Fatalf("len(ledger.Entries) = %d, want 3", len(ledger.Entries))
	}

	scheduledCount := 0
	deferredCount := 0
	deferredRuns := int32(0)
	var minStart time.Time
	var maxFinish time.Time

	for _, e := range ledger.Entries {
		if e.StartedAt == "" {
			deferredCount++
			deferredRuns += e.Runs
			if e.FinishedAt != "" {
				t.Fatalf("deferred job should not have finished_at, got %q", e.FinishedAt)
			}
			continue
		}

		scheduledCount++
		start, err := time.Parse(time.RFC3339, e.StartedAt)
		if err != nil {
			t.Fatalf("parse started_at: %v", err)
		}
		finish, err := time.Parse(time.RFC3339, e.FinishedAt)
		if err != nil {
			t.Fatalf("parse finished_at: %v", err)
		}
		if minStart.IsZero() || start.Before(minStart) {
			minStart = start
		}
		if maxFinish.IsZero() || finish.After(maxFinish) {
			maxFinish = finish
		}
		if e.Runs <= 0 || e.Runs > 2 {
			t.Fatalf("scheduled chunk runs = %d, want 1..2", e.Runs)
		}
	}

	if scheduledCount != 2 {
		t.Fatalf("scheduledCount = %d, want 2", scheduledCount)
	}
	if deferredCount != 1 {
		t.Fatalf("deferredCount = %d, want 1", deferredCount)
	}
	if deferredRuns != 6 {
		t.Fatalf("deferredRuns = %d, want 6", deferredRuns)
	}
	if maxFinish.Sub(minStart) > 2*24*time.Hour {
		t.Fatalf("scheduled horizon span = %s, want <= 48h", maxFinish.Sub(minStart))
	}
}

func TestIndustryLedgerGetProjectSnapshotIncludesMaterialDiff(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	project, err := d.CreateIndustryProjectForUser("user-snapshot", IndustryProjectCreateInput{
		Name: "Snapshot Coverage Project",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	_, err = d.ApplyIndustryPlanForUser("user-snapshot", project.ID, IndustryPlanPatch{
		Replace: true,
		Tasks: []IndustryTaskPlanInput{
			{
				Name:       "Task A",
				Activity:   "manufacturing",
				TargetRuns: 2,
			},
		},
		Jobs: []IndustryJobPlanInput{
			{
				Activity:        "manufacturing",
				Runs:            2,
				DurationSeconds: 3600,
				CostISK:         250000,
				Status:          IndustryJobStatusQueued,
			},
		},
		Materials: []IndustryMaterialPlanInput{
			{
				TaskID:       0,
				TypeID:       34,
				TypeName:     "Tritanium",
				RequiredQty:  100,
				AvailableQty: 20,
				BuyQty:       50,
				BuildQty:     10,
				UnitCostISK:  5.1,
				Source:       "market",
			},
			{
				TaskID:       0,
				TypeID:       35,
				TypeName:     "Pyerite",
				RequiredQty:  40,
				AvailableQty: 10,
				BuyQty:       30,
				BuildQty:     0,
				UnitCostISK:  9.5,
				Source:       "stock",
			},
		},
		Blueprints: []IndustryBlueprintPoolInput{
			{
				BlueprintTypeID: 910001,
				BlueprintName:   "Snapshot Test Blueprint",
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

	snapshot, err := d.GetIndustryProjectSnapshotForUser("user-snapshot", project.ID)
	if err != nil {
		t.Fatalf("GetIndustryProjectSnapshotForUser: %v", err)
	}

	if snapshot.Project.ID != project.ID {
		t.Fatalf("snapshot project id = %d, want %d", snapshot.Project.ID, project.ID)
	}
	if len(snapshot.Tasks) != 1 {
		t.Fatalf("snapshot tasks len = %d, want 1", len(snapshot.Tasks))
	}
	if len(snapshot.Jobs) != 1 {
		t.Fatalf("snapshot jobs len = %d, want 1", len(snapshot.Jobs))
	}
	if len(snapshot.Materials) != 2 {
		t.Fatalf("snapshot materials len = %d, want 2", len(snapshot.Materials))
	}
	if len(snapshot.Blueprints) != 1 {
		t.Fatalf("snapshot blueprints len = %d, want 1", len(snapshot.Blueprints))
	}
	if len(snapshot.MaterialDiff) != 2 {
		t.Fatalf("snapshot material_diff len = %d, want 2", len(snapshot.MaterialDiff))
	}

	byType := map[int32]IndustryMaterialDiff{}
	for _, diff := range snapshot.MaterialDiff {
		byType[diff.TypeID] = diff
	}
	if got := byType[34].MissingQty; got != 20 {
		t.Fatalf("type 34 missing_qty = %d, want 20", got)
	}
	if got := byType[35].MissingQty; got != 0 {
		t.Fatalf("type 35 missing_qty = %d, want 0", got)
	}
}

func TestIndustryLedgerRebalanceLocationFirstAvoidsDoubleAllocation(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	project, err := d.CreateIndustryProjectForUser("user-rebalance", IndustryProjectCreateInput{
		Name: "Rebalance Reservation Test",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	_, err = d.ApplyIndustryPlanForUser("user-rebalance", project.ID, IndustryPlanPatch{
		Replace: true,
		Tasks: []IndustryTaskPlanInput{
			{
				Name:       "Task A",
				Activity:   "manufacturing",
				TargetRuns: 1,
				Constraints: json.RawMessage(`{
					"warehouse_location_id": 1001
				}`),
			},
			{
				Name:       "Task B",
				Activity:   "manufacturing",
				TargetRuns: 1,
				Constraints: json.RawMessage(`{
					"warehouse_location_id": 1002
				}`),
			},
		},
	})
	if err != nil {
		t.Fatalf("ApplyIndustryPlanForUser tasks: %v", err)
	}

	snapshot, err := d.GetIndustryProjectSnapshotForUser("user-rebalance", project.ID)
	if err != nil {
		t.Fatalf("GetIndustryProjectSnapshotForUser: %v", err)
	}
	taskIDByName := map[string]int64{}
	for _, task := range snapshot.Tasks {
		taskIDByName[task.Name] = task.ID
	}
	taskAID := taskIDByName["Task A"]
	taskBID := taskIDByName["Task B"]
	if taskAID <= 0 || taskBID <= 0 {
		t.Fatalf("failed to resolve task IDs: %+v", taskIDByName)
	}

	_, err = d.ApplyIndustryPlanForUser("user-rebalance", project.ID, IndustryPlanPatch{
		Replace: false,
		Materials: []IndustryMaterialPlanInput{
			{
				TaskID:       taskAID,
				TypeID:       34,
				TypeName:     "Tritanium",
				RequiredQty:  80,
				AvailableQty: 0,
				BuyQty:       80,
				BuildQty:     0,
				Source:       "market",
			},
			{
				TaskID:       taskBID,
				TypeID:       34,
				TypeName:     "Tritanium",
				RequiredQty:  50,
				AvailableQty: 0,
				BuyQty:       50,
				BuildQty:     0,
				Source:       "market",
			},
		},
	})
	if err != nil {
		t.Fatalf("ApplyIndustryPlanForUser materials: %v", err)
	}

	updated, err := d.RebalanceIndustryProjectMaterialsFromStockForUser(
		"user-rebalance",
		project.ID,
		map[int32]int64{
			34: 100,
		},
		map[int32]map[int64]int64{
			34: {
				1001: 50,
				1002: 50,
			},
		},
		"location_first",
		"buy",
	)
	if err != nil {
		t.Fatalf("RebalanceIndustryProjectMaterialsFromStockForUser: %v", err)
	}
	if len(updated) != 2 {
		t.Fatalf("len(updated) = %d, want 2", len(updated))
	}

	byTaskID := map[int64]IndustryMaterialPlan{}
	var totalAvailable int64
	for _, row := range updated {
		byTaskID[row.TaskID] = row
		totalAvailable += row.AvailableQty
	}

	if totalAvailable != 100 {
		t.Fatalf("total available = %d, want 100", totalAvailable)
	}
	if got := byTaskID[taskAID].AvailableQty; got != 80 {
		t.Fatalf("task A available_qty = %d, want 80", got)
	}
	if got := byTaskID[taskAID].BuyQty; got != 0 {
		t.Fatalf("task A buy_qty = %d, want 0", got)
	}
	if got := byTaskID[taskBID].AvailableQty; got != 20 {
		t.Fatalf("task B available_qty = %d, want 20", got)
	}
	if got := byTaskID[taskBID].BuyQty; got != 30 {
		t.Fatalf("task B buy_qty = %d, want 30", got)
	}
}

func TestIndustryLedgerRebalanceLocationInventoryCappedByTotal(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	project, err := d.CreateIndustryProjectForUser("user-rebalance-cap", IndustryProjectCreateInput{
		Name: "Rebalance Cap Test",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	_, err = d.ApplyIndustryPlanForUser("user-rebalance-cap", project.ID, IndustryPlanPatch{
		Replace: true,
		Tasks: []IndustryTaskPlanInput{
			{
				Name:       "Task Cap",
				Activity:   "manufacturing",
				TargetRuns: 1,
				Constraints: json.RawMessage(`{
					"warehouse_location_id": 1001
				}`),
			},
		},
	})
	if err != nil {
		t.Fatalf("ApplyIndustryPlanForUser tasks: %v", err)
	}

	snapshot, err := d.GetIndustryProjectSnapshotForUser("user-rebalance-cap", project.ID)
	if err != nil {
		t.Fatalf("GetIndustryProjectSnapshotForUser: %v", err)
	}
	if len(snapshot.Tasks) != 1 {
		t.Fatalf("snapshot tasks len = %d, want 1", len(snapshot.Tasks))
	}
	taskID := snapshot.Tasks[0].ID

	_, err = d.ApplyIndustryPlanForUser("user-rebalance-cap", project.ID, IndustryPlanPatch{
		Replace: false,
		Materials: []IndustryMaterialPlanInput{
			{
				TaskID:       taskID,
				TypeID:       34,
				TypeName:     "Tritanium",
				RequiredQty:  60,
				AvailableQty: 0,
				BuyQty:       60,
				BuildQty:     0,
				Source:       "market",
			},
		},
	})
	if err != nil {
		t.Fatalf("ApplyIndustryPlanForUser materials: %v", err)
	}

	updated, err := d.RebalanceIndustryProjectMaterialsFromStockForUser(
		"user-rebalance-cap",
		project.ID,
		map[int32]int64{
			34: 40, // authoritative total
		},
		map[int32]map[int64]int64{
			34: {
				1001: 80, // inconsistent location qty should be capped to total
			},
		},
		"location_first",
		"buy",
	)
	if err != nil {
		t.Fatalf("RebalanceIndustryProjectMaterialsFromStockForUser: %v", err)
	}
	if len(updated) != 1 {
		t.Fatalf("len(updated) = %d, want 1", len(updated))
	}
	row := updated[0]
	if row.AvailableQty != 40 {
		t.Fatalf("available_qty = %d, want 40", row.AvailableQty)
	}
	if row.BuyQty != 20 {
		t.Fatalf("buy_qty = %d, want 20", row.BuyQty)
	}
}

func TestIndustryLedgerStrictBlueprintGateBlocksMissingPool(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	project, err := d.CreateIndustryProjectForUser("user-bp-gate", IndustryProjectCreateInput{
		Name: "Blueprint Gate Strict",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	summary, err := d.ApplyIndustryPlanForUser("user-bp-gate", project.ID, IndustryPlanPatch{
		Replace: true,
		Tasks: []IndustryTaskPlanInput{
			{
				Name:       "Strict BP Task",
				Activity:   "manufacturing",
				TargetRuns: 5,
				Constraints: json.RawMessage(`{
					"blueprint_type_id": 910001,
					"blueprint_location_id": 60003760,
					"duration_seconds_per_run": 60
				}`),
			},
		},
		Scheduler: IndustryPlanSchedulerInput{
			Enabled: true,
		},
	})
	if err != nil {
		t.Fatalf("ApplyIndustryPlanForUser: %v", err)
	}
	if summary.JobsInserted != 0 {
		t.Fatalf("summary.JobsInserted = %d, want 0", summary.JobsInserted)
	}
	foundGateWarning := false
	for _, warning := range summary.Warnings {
		if strings.Contains(warning, "blueprint gate:") {
			foundGateWarning = true
			break
		}
	}
	if !foundGateWarning {
		t.Fatalf("expected blueprint gate warning, got: %v", summary.Warnings)
	}

	ledger, err := d.GetIndustryLedgerForUser("user-bp-gate", IndustryLedgerOptions{
		ProjectID: project.ID,
		Limit:     20,
	})
	if err != nil {
		t.Fatalf("GetIndustryLedgerForUser: %v", err)
	}
	if ledger.Total != 0 {
		t.Fatalf("ledger.Total = %d, want 0", ledger.Total)
	}
}

func TestIndustryLedgerStrictBlueprintGateBypassAllowsMissingPool(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	project, err := d.CreateIndustryProjectForUser("user-bp-bypass", IndustryProjectCreateInput{
		Name: "Blueprint Gate Bypass",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	summary, err := d.ApplyIndustryPlanForUser("user-bp-bypass", project.ID, IndustryPlanPatch{
		Replace: true,
		Tasks: []IndustryTaskPlanInput{
			{
				Name:       "Bypass BP Task",
				Activity:   "manufacturing",
				TargetRuns: 3,
				Constraints: json.RawMessage(`{
					"blueprint_type_id": 910002,
					"blueprint_location_id": 60003760,
					"duration_seconds_per_run": 60
				}`),
			},
		},
		Scheduler: IndustryPlanSchedulerInput{
			Enabled: true,
		},
		StrictBPBypass: true,
	})
	if err != nil {
		t.Fatalf("ApplyIndustryPlanForUser: %v", err)
	}
	if summary.JobsInserted == 0 {
		t.Fatalf("summary.JobsInserted = %d, want > 0 with bypass", summary.JobsInserted)
	}

	ledger, err := d.GetIndustryLedgerForUser("user-bp-bypass", IndustryLedgerOptions{
		ProjectID: project.ID,
		Limit:     20,
	})
	if err != nil {
		t.Fatalf("GetIndustryLedgerForUser: %v", err)
	}
	if ledger.Total == 0 {
		t.Fatalf("ledger.Total = %d, want > 0 with bypass", ledger.Total)
	}
}

func TestIndustryLedgerStrictBlueprintGateLocationMismatch(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	project, err := d.CreateIndustryProjectForUser("user-bp-location", IndustryProjectCreateInput{
		Name: "Blueprint Location Gate",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	summary, err := d.ApplyIndustryPlanForUser("user-bp-location", project.ID, IndustryPlanPatch{
		Replace: true,
		Tasks: []IndustryTaskPlanInput{
			{
				Name:       "Location Locked Task",
				Activity:   "manufacturing",
				TargetRuns: 4,
				Constraints: json.RawMessage(`{
					"blueprint_type_id": 910003,
					"blueprint_location_id": 70000001,
					"duration_seconds_per_run": 60
				}`),
			},
		},
		Blueprints: []IndustryBlueprintPoolInput{
			{
				BlueprintTypeID: 910003,
				BlueprintName:   "Mismatch Blueprint",
				LocationID:      60003760, // different location
				Quantity:        1,
				IsBPO:           false,
				AvailableRuns:   10,
			},
		},
		Scheduler: IndustryPlanSchedulerInput{
			Enabled: true,
		},
	})
	if err != nil {
		t.Fatalf("ApplyIndustryPlanForUser: %v", err)
	}
	if summary.JobsInserted != 0 {
		t.Fatalf("summary.JobsInserted = %d, want 0", summary.JobsInserted)
	}
	foundLocationWarning := false
	for _, warning := range summary.Warnings {
		if strings.Contains(warning, "missing blueprint at location") {
			foundLocationWarning = true
			break
		}
	}
	if !foundLocationWarning {
		t.Fatalf("expected location mismatch warning, got: %v", summary.Warnings)
	}
}

func TestIndustryLedgerApplyPlanRemapsJobTaskIDWithoutScheduler(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	seedProject, err := d.CreateIndustryProjectForUser("user-job-remap", IndustryProjectCreateInput{
		Name: "Seed Project",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser seed: %v", err)
	}
	if _, err := d.ApplyIndustryPlanForUser("user-job-remap", seedProject.ID, IndustryPlanPatch{
		Replace: true,
		Tasks: []IndustryTaskPlanInput{
			{Name: "Seed task", Activity: "manufacturing", TargetRuns: 1},
		},
	}); err != nil {
		t.Fatalf("ApplyIndustryPlanForUser seed: %v", err)
	}

	project, err := d.CreateIndustryProjectForUser("user-job-remap", IndustryProjectCreateInput{
		Name: "Remap Target",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser target: %v", err)
	}

	if _, err := d.ApplyIndustryPlanForUser("user-job-remap", project.ID, IndustryPlanPatch{
		Replace: true,
		Tasks: []IndustryTaskPlanInput{
			{Name: "Root Task", Activity: "manufacturing", TargetRuns: 2},
			{Name: "Child Task", Activity: "manufacturing", TargetRuns: 1, ParentTaskID: 1},
		},
		Jobs: []IndustryJobPlanInput{
			{
				TaskID:          1, // input index reference, should map to inserted Root Task ID
				Activity:        "manufacturing",
				Runs:            2,
				DurationSeconds: 600,
				CostISK:         1000,
				Status:          IndustryJobStatusPlanned,
			},
		},
	}); err != nil {
		t.Fatalf("ApplyIndustryPlanForUser target: %v", err)
	}

	snapshot, err := d.GetIndustryProjectSnapshotForUser("user-job-remap", project.ID)
	if err != nil {
		t.Fatalf("GetIndustryProjectSnapshotForUser: %v", err)
	}
	if len(snapshot.Tasks) != 2 {
		t.Fatalf("len(snapshot.Tasks) = %d, want 2", len(snapshot.Tasks))
	}
	if len(snapshot.Jobs) != 1 {
		t.Fatalf("len(snapshot.Jobs) = %d, want 1", len(snapshot.Jobs))
	}

	var rootTaskID int64
	for _, task := range snapshot.Tasks {
		if task.Name == "Root Task" {
			rootTaskID = task.ID
			break
		}
	}
	if rootTaskID <= 0 {
		t.Fatalf("root task not found in snapshot: %+v", snapshot.Tasks)
	}
	if snapshot.Jobs[0].TaskID != rootTaskID {
		t.Fatalf("job.task_id = %d, want %d (root task id)", snapshot.Jobs[0].TaskID, rootTaskID)
	}
}

func TestIndustryLedgerAppendPlanTaskReferencesPreferExistingIDAndSupportNegativeRowRefs(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	project, err := d.CreateIndustryProjectForUser("user-job-remap-append", IndustryProjectCreateInput{
		Name: "Remap Append Target",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	if _, err := d.ApplyIndustryPlanForUser("user-job-remap-append", project.ID, IndustryPlanPatch{
		Replace: true,
		Tasks: []IndustryTaskPlanInput{
			{Name: "Existing Root", Activity: "manufacturing", TargetRuns: 1},
		},
	}); err != nil {
		t.Fatalf("ApplyIndustryPlanForUser seed: %v", err)
	}

	seedSnapshot, err := d.GetIndustryProjectSnapshotForUser("user-job-remap-append", project.ID)
	if err != nil {
		t.Fatalf("GetIndustryProjectSnapshotForUser seed: %v", err)
	}
	if len(seedSnapshot.Tasks) != 1 {
		t.Fatalf("len(seedSnapshot.Tasks) = %d, want 1", len(seedSnapshot.Tasks))
	}
	existingTaskID := seedSnapshot.Tasks[0].ID
	if existingTaskID <= 0 {
		t.Fatalf("existing task id = %d, want > 0", existingTaskID)
	}

	summary, err := d.ApplyIndustryPlanForUser("user-job-remap-append", project.ID, IndustryPlanPatch{
		Replace: false,
		Tasks: []IndustryTaskPlanInput{
			{Name: "Appended Root", Activity: "manufacturing", TargetRuns: 2},
		},
		Jobs: []IndustryJobPlanInput{
			{
				TaskID:          existingTaskID, // ambiguous in append mode when existingTaskID==1
				Activity:        "manufacturing",
				Runs:            1,
				DurationSeconds: 600,
				CostISK:         1000,
				Status:          IndustryJobStatusPlanned,
				Notes:           "link-existing",
			},
			{
				TaskID:          -1, // explicit row reference to appended task row #1
				Activity:        "manufacturing",
				Runs:            1,
				DurationSeconds: 600,
				CostISK:         1000,
				Status:          IndustryJobStatusPlanned,
				Notes:           "link-row",
			},
		},
	})
	if err != nil {
		t.Fatalf("ApplyIndustryPlanForUser append: %v", err)
	}
	if summary.TasksInserted != 1 {
		t.Fatalf("summary.TasksInserted = %d, want 1", summary.TasksInserted)
	}
	if summary.JobsInserted != 2 {
		t.Fatalf("summary.JobsInserted = %d, want 2", summary.JobsInserted)
	}
	foundAmbiguityWarning := false
	for _, warning := range summary.Warnings {
		if strings.Contains(warning, industryTaskRefAmbiguityWarning) {
			foundAmbiguityWarning = true
			break
		}
	}
	if existingTaskID == 1 && !foundAmbiguityWarning {
		t.Fatalf("expected ambiguity warning when existingTaskID=1, got: %v", summary.Warnings)
	}

	snapshot, err := d.GetIndustryProjectSnapshotForUser("user-job-remap-append", project.ID)
	if err != nil {
		t.Fatalf("GetIndustryProjectSnapshotForUser: %v", err)
	}
	if len(snapshot.Tasks) != 2 {
		t.Fatalf("len(snapshot.Tasks) = %d, want 2", len(snapshot.Tasks))
	}
	if len(snapshot.Jobs) != 2 {
		t.Fatalf("len(snapshot.Jobs) = %d, want 2", len(snapshot.Jobs))
	}

	var appendedTaskID int64
	for _, task := range snapshot.Tasks {
		if task.Name == "Appended Root" {
			appendedTaskID = task.ID
			break
		}
	}
	if appendedTaskID <= 0 {
		t.Fatalf("appended task not found in snapshot: %+v", snapshot.Tasks)
	}

	var existingLinkedTaskID int64
	var rowLinkedTaskID int64
	for _, job := range snapshot.Jobs {
		switch job.Notes {
		case "link-existing":
			existingLinkedTaskID = job.TaskID
		case "link-row":
			rowLinkedTaskID = job.TaskID
		}
	}
	if existingLinkedTaskID != existingTaskID {
		t.Fatalf("existing-linked job.task_id = %d, want existing task id %d", existingLinkedTaskID, existingTaskID)
	}
	if rowLinkedTaskID != appendedTaskID {
		t.Fatalf("row-linked job.task_id = %d, want appended task id %d", rowLinkedTaskID, appendedTaskID)
	}
}

func TestIndustryLedgerPreviewAppendPlanSupportsNegativeRowRefs(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	project, err := d.CreateIndustryProjectForUser("user-job-remap-preview", IndustryProjectCreateInput{
		Name: "Remap Preview Target",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}
	if _, err := d.ApplyIndustryPlanForUser("user-job-remap-preview", project.ID, IndustryPlanPatch{
		Replace: true,
		Tasks: []IndustryTaskPlanInput{
			{Name: "Existing Root", Activity: "manufacturing", TargetRuns: 1},
		},
	}); err != nil {
		t.Fatalf("ApplyIndustryPlanForUser seed: %v", err)
	}

	seedSnapshot, err := d.GetIndustryProjectSnapshotForUser("user-job-remap-preview", project.ID)
	if err != nil {
		t.Fatalf("GetIndustryProjectSnapshotForUser seed: %v", err)
	}
	if len(seedSnapshot.Tasks) != 1 {
		t.Fatalf("len(seedSnapshot.Tasks) = %d, want 1", len(seedSnapshot.Tasks))
	}
	existingTaskID := seedSnapshot.Tasks[0].ID

	preview, err := d.PreviewIndustryPlanForUser("user-job-remap-preview", project.ID, IndustryPlanPatch{
		Replace: false,
		Tasks: []IndustryTaskPlanInput{
			{Name: "Preview Root", Activity: "manufacturing", TargetRuns: 2},
		},
		Jobs: []IndustryJobPlanInput{
			{
				TaskID:          existingTaskID,
				Activity:        "manufacturing",
				Runs:            1,
				DurationSeconds: 600,
				CostISK:         1000,
				Status:          IndustryJobStatusPlanned,
				Notes:           "preview-existing",
			},
			{
				TaskID:          -1,
				Activity:        "manufacturing",
				Runs:            1,
				DurationSeconds: 600,
				CostISK:         1000,
				Status:          IndustryJobStatusPlanned,
				Notes:           "preview-row",
			},
		},
	})
	if err != nil {
		t.Fatalf("PreviewIndustryPlanForUser: %v", err)
	}
	if len(preview.Tasks) != 1 {
		t.Fatalf("len(preview.Tasks) = %d, want 1", len(preview.Tasks))
	}
	if len(preview.Jobs) != 2 {
		t.Fatalf("len(preview.Jobs) = %d, want 2", len(preview.Jobs))
	}

	previewTaskID := preview.Tasks[0].TaskID
	if previewTaskID <= 0 {
		t.Fatalf("preview task id = %d, want > 0", previewTaskID)
	}

	var existingLinkedTaskID int64
	var rowLinkedTaskID int64
	for _, job := range preview.Jobs {
		switch job.Notes {
		case "preview-existing":
			existingLinkedTaskID = job.TaskID
		case "preview-row":
			rowLinkedTaskID = job.TaskID
		}
	}
	if existingLinkedTaskID != existingTaskID {
		t.Fatalf("preview existing-linked task_id = %d, want %d", existingLinkedTaskID, existingTaskID)
	}
	if rowLinkedTaskID != previewTaskID {
		t.Fatalf("preview row-linked task_id = %d, want %d", rowLinkedTaskID, previewTaskID)
	}
}

func TestIndustryLedgerJobStatusUpdateValidatesRFC3339(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	project, err := d.CreateIndustryProjectForUser("user-job-rfc3339", IndustryProjectCreateInput{
		Name: "Job RFC3339 Validation",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}
	if _, err := d.ApplyIndustryPlanForUser("user-job-rfc3339", project.ID, IndustryPlanPatch{
		Replace: true,
		Jobs: []IndustryJobPlanInput{
			{
				Activity:        "manufacturing",
				Runs:            1,
				DurationSeconds: 1200,
				CostISK:         250000,
				Status:          IndustryJobStatusPlanned,
			},
		},
	}); err != nil {
		t.Fatalf("ApplyIndustryPlanForUser: %v", err)
	}

	ledger, err := d.GetIndustryLedgerForUser("user-job-rfc3339", IndustryLedgerOptions{
		ProjectID: project.ID,
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("GetIndustryLedgerForUser: %v", err)
	}
	if len(ledger.Entries) != 1 {
		t.Fatalf("len(ledger.Entries) = %d, want 1", len(ledger.Entries))
	}
	jobID := ledger.Entries[0].JobID

	if _, err := d.UpdateIndustryJobStatusForUser(
		"user-job-rfc3339",
		jobID,
		IndustryJobStatusActive,
		"2026-02-24 10:00:00",
		"",
		"",
	); err == nil || !strings.Contains(strings.ToLower(err.Error()), "started_at must be rfc3339") {
		t.Fatalf("expected started_at RFC3339 validation error, got: %v", err)
	}

	if _, err := d.UpdateIndustryJobStatusesForUser(
		"user-job-rfc3339",
		[]int64{jobID},
		IndustryJobStatusCompleted,
		"",
		"not-a-timestamp",
		"",
	); err == nil || !strings.Contains(strings.ToLower(err.Error()), "finished_at must be rfc3339") {
		t.Fatalf("expected finished_at RFC3339 validation error, got: %v", err)
	}
}

func TestIndustryLedgerReplaceBlueprintPoolDoesNotTouchTasksOrJobs(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	project, err := d.CreateIndustryProjectForUser("user-bp-replace", IndustryProjectCreateInput{
		Name: "Blueprint Replace Scope",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	if _, err := d.ApplyIndustryPlanForUser("user-bp-replace", project.ID, IndustryPlanPatch{
		Replace: true,
		Tasks: []IndustryTaskPlanInput{
			{Name: "Task baseline", Activity: "manufacturing", TargetRuns: 1},
		},
		Jobs: []IndustryJobPlanInput{
			{Activity: "manufacturing", Runs: 1, DurationSeconds: 300, CostISK: 10000, Status: IndustryJobStatusPlanned},
		},
		Blueprints: []IndustryBlueprintPoolInput{
			{BlueprintTypeID: 910100, BlueprintName: "Legacy BP", IsBPO: true, AvailableRuns: 0, Quantity: 1},
		},
	}); err != nil {
		t.Fatalf("ApplyIndustryPlanForUser baseline: %v", err)
	}

	if _, err := d.ApplyIndustryPlanForUser("user-bp-replace", project.ID, IndustryPlanPatch{
		ReplaceBlueprintPool: true,
		Blueprints: []IndustryBlueprintPoolInput{
			{BlueprintTypeID: 910200, BlueprintName: "Fresh BP", IsBPO: true, AvailableRuns: 0, Quantity: 1},
		},
	}); err != nil {
		t.Fatalf("ApplyIndustryPlanForUser replace blueprints: %v", err)
	}

	snapshot, err := d.GetIndustryProjectSnapshotForUser("user-bp-replace", project.ID)
	if err != nil {
		t.Fatalf("GetIndustryProjectSnapshotForUser: %v", err)
	}
	if len(snapshot.Tasks) != 1 {
		t.Fatalf("len(snapshot.Tasks) = %d, want 1", len(snapshot.Tasks))
	}
	if len(snapshot.Jobs) != 1 {
		t.Fatalf("len(snapshot.Jobs) = %d, want 1", len(snapshot.Jobs))
	}
	if len(snapshot.Blueprints) != 1 {
		t.Fatalf("len(snapshot.Blueprints) = %d, want 1", len(snapshot.Blueprints))
	}
	if snapshot.Blueprints[0].BlueprintTypeID != 910200 {
		t.Fatalf("blueprint_type_id = %d, want 910200", snapshot.Blueprints[0].BlueprintTypeID)
	}
}

func TestIndustryLedgerBPOAvailableRunsAlwaysNormalizedToZero(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	project, err := d.CreateIndustryProjectForUser("user-bpo-normalize", IndustryProjectCreateInput{
		Name: "BPO Runs Normalize",
	})
	if err != nil {
		t.Fatalf("CreateIndustryProjectForUser: %v", err)
	}

	if _, err := d.ApplyIndustryPlanForUser("user-bpo-normalize", project.ID, IndustryPlanPatch{
		Replace: true,
		Blueprints: []IndustryBlueprintPoolInput{
			{
				BlueprintTypeID: 910300,
				BlueprintName:   "Unlimited BPO",
				IsBPO:           true,
				Quantity:        1,
				AvailableRuns:   9999, // should be ignored for BPO
			},
		},
	}); err != nil {
		t.Fatalf("ApplyIndustryPlanForUser: %v", err)
	}

	snapshot, err := d.GetIndustryProjectSnapshotForUser("user-bpo-normalize", project.ID)
	if err != nil {
		t.Fatalf("GetIndustryProjectSnapshotForUser: %v", err)
	}
	if len(snapshot.Blueprints) != 1 {
		t.Fatalf("len(snapshot.Blueprints) = %d, want 1", len(snapshot.Blueprints))
	}
	if snapshot.Blueprints[0].AvailableRuns != 0 {
		t.Fatalf("bpo available_runs = %d, want 0", snapshot.Blueprints[0].AvailableRuns)
	}
}
