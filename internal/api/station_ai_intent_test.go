package api

import (
	"strings"
	"testing"
)

func TestDetectStationAIIntent(t *testing.T) {
	tests := []struct {
		name    string
		msg     string
		history []stationAIHistoryMessage
		want    stationAIIntentKind
	}{
		{
			name: "smalltalk greeting",
			msg:  "привет",
			want: stationAIIntentSmallTalk,
		},
		{
			name: "trading request",
			msg:  "покажи топ сделки в текущем скане",
			want: stationAIIntentTrading,
		},
		{
			name: "debug request",
			msg:  "у меня ошибка undefined в чате",
			want: stationAIIntentDebug,
		},
		{
			name: "product help request",
			msg:  "как работает station trading в проекте",
			want: stationAIIntentProduct,
		},
		{
			name: "web research request",
			msg:  "погугли свежие новости по eve market",
			want: stationAIIntentResearch,
		},
		{
			name: "followup why after trading answer",
			msg:  "почему?",
			history: []stationAIHistoryMessage{
				{Role: "assistant", Content: "Recommendation: adjust filter and check risk"},
			},
			want: stationAIIntentTrading,
		},
		{
			name: "strict json trading prompt",
			msg:  "Use ONLY the JSON context scan_snapshot summary rows runtime. Build decision matrix execute now monitor reject and capital allocation.",
			want: stationAIIntentTrading,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := detectStationAIIntent(tc.msg, tc.history)
			if got != tc.want {
				t.Fatalf("intent mismatch: got=%q want=%q", got, tc.want)
			}
		})
	}
}

func TestNormalizeStationAIHistory(t *testing.T) {
	input := []stationAIHistoryMessage{
		{Role: "user", Content: "  one  "},
		{Role: "assistant", Content: "two"},
		{Role: "system", Content: "ignore"},
		{Role: "user", Content: ""},
	}
	out := normalizeStationAIHistory(input)
	if len(out) != 2 {
		t.Fatalf("expected 2 history messages, got %d", len(out))
	}
	if out[0].Content != "one" || out[1].Content != "two" {
		t.Fatalf("unexpected normalized content: %#v", out)
	}
}

func TestNormalizeStationAIHistoryDropsDiagnosticAssistantMessages(t *testing.T) {
	input := []stationAIHistoryMessage{
		{Role: "user", Content: "analyze scan"},
		{Role: "assistant", Content: `{"status":"NEED_FULL_CONTEXT","rows_seen_count":0}`},
		{Role: "assistant", Content: "Recommendation: keep min_margin 10 and min_daily_volume 12."},
	}
	out := normalizeStationAIHistory(input)
	if len(out) != 2 {
		t.Fatalf("expected 2 history messages after diagnostic filtering, got %d", len(out))
	}
	if out[0].Role != "user" || out[1].Role != "assistant" {
		t.Fatalf("unexpected role sequence after filtering: %#v", out)
	}
	if strings.Contains(strings.ToLower(out[1].Content), "need_full_context") {
		t.Fatalf("diagnostic assistant content should be filtered: %#v", out[1])
	}
}

func TestStationAIContextForIntentSmallTalk(t *testing.T) {
	ctx := stationAIContextPayload{
		Rows: []stationAIContextRow{
			{TypeID: 1, TypeName: "Tritanium"},
		},
		Summary: stationAIContextSummary{
			TotalRows: 1,
		},
	}
	got := stationAIContextForIntent(ctx, stationAIIntentSmallTalk)
	if len(got.Rows) != 0 {
		t.Fatalf("expected empty rows for smalltalk, got %d", len(got.Rows))
	}
	if got.Summary.TotalRows != 0 {
		t.Fatalf("expected zero summary for smalltalk, got %#v", got.Summary)
	}
}

func TestStationAIRequestsFullScanSettings(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{
			name: "ru full list",
			msg:  "Напиши полный список настроек сканирования, включая обычные и адвансед поля",
			want: true,
		},
		{
			name: "en full list",
			msg:  "Please give me full list of scan settings with all fields",
			want: true,
		},
		{
			name: "explicit phrase",
			msg:  "list all scan parameters",
			want: true,
		},
		{
			name: "regular recommendation request",
			msg:  "what do you think about current station trading results",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := stationAIRequestsFullScanSettings(tc.msg)
			if got != tc.want {
				t.Fatalf("full settings detect mismatch: got=%v want=%v msg=%q", got, tc.want, tc.msg)
			}
		})
	}
}

func TestStationAIUserPromptIncludesFullSettingsDirective(t *testing.T) {
	plan := stationAIPlannerPlan{
		Intent:       stationAIIntentTrading,
		ContextLevel: "full",
		ResponseMode: "structured",
		NeedWiki:     false,
		NeedWeb:      false,
	}
	ctxJSON := []byte(`{"scan_snapshot":{"radius":15,"min_margin":8}}`)
	prompt := stationAIUserPrompt(
		"ru",
		"дай полный список параметров сканирования",
		ctxJSON,
		plan,
	)
	if !strings.Contains(prompt, "ВСЕХ полей объекта scan_snapshot") {
		t.Fatalf("expected full settings directive in prompt, got: %s", prompt)
	}
}

func TestStationAIWebQueryVariants(t *testing.T) {
	variants := stationAIWebQueryVariants(
		"ru",
		"что нового по рынку eve online и station trading",
		stationAIIntentTrading,
	)
	if len(variants) == 0 {
		t.Fatal("expected non-empty query variants")
	}
	if len(variants) > stationAIWebMaxQueries {
		t.Fatalf("too many query variants: got=%d max=%d", len(variants), stationAIWebMaxQueries)
	}
	if !strings.Contains(strings.ToLower(variants[0]), "рынку") &&
		!strings.Contains(strings.ToLower(variants[0]), "market") {
		t.Fatalf("expected first query to preserve user message context, got=%q", variants[0])
	}

	foundEveHint := false
	for _, q := range variants {
		if strings.Contains(strings.ToLower(q), "eve online") {
			foundEveHint = true
			break
		}
	}
	if !foundEveHint {
		t.Fatalf("expected at least one variant with eve online hint, got=%v", variants)
	}
}

func TestStationAINeedsRuntimeContext(t *testing.T) {
	tests := []struct {
		name   string
		intent stationAIIntentKind
		msg    string
		want   bool
	}{
		{
			name:   "wallet balance request",
			intent: stationAIIntentTrading,
			msg:    "show my wallet balance and pnl",
			want:   true,
		},
		{
			name:   "ru transactions request",
			intent: stationAIIntentTrading,
			msg:    "покажи мои транзакции и риск за месяц",
			want:   true,
		},
		{
			name:   "generic scan question",
			intent: stationAIIntentTrading,
			msg:    "что думаешь о текущих результатах station trading",
			want:   false,
		},
		{
			name:   "product docs question",
			intent: stationAIIntentProduct,
			msg:    "как работает radius scan",
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := stationAINeedsRuntimeContext(tc.intent, tc.msg)
			if got != tc.want {
				t.Fatalf("runtime context detect mismatch: got=%v want=%v msg=%q", got, tc.want, tc.msg)
			}
		})
	}
}

func TestStationAIPreflightFailOnMissingRows(t *testing.T) {
	plan := stationAIPlannerPlan{
		Intent:       stationAIIntentTrading,
		ContextLevel: "full",
	}
	ctx := stationAIContextPayload{
		Summary: stationAIContextSummary{
			VisibleRows: 0,
		},
	}
	got := stationAIPreflight("en", plan, ctx, false)
	if got.Status != "fail" {
		t.Fatalf("expected fail preflight status, got %q", got.Status)
	}
	if len(got.Missing) == 0 {
		t.Fatalf("expected missing fields in preflight result, got %#v", got)
	}
}

func TestStationAIPreflightPartialWhenRuntimeUnavailable(t *testing.T) {
	plan := stationAIPlannerPlan{
		Intent:       stationAIIntentGeneral,
		ContextLevel: "summary",
	}
	ctx := stationAIContextPayload{
		Summary: stationAIContextSummary{
			VisibleRows: 5,
		},
	}
	got := stationAIPreflight("en", plan, ctx, true)
	if got.Status != "partial" {
		t.Fatalf("expected partial preflight status, got %q", got.Status)
	}
	if len(got.Caveats) == 0 {
		t.Fatalf("expected runtime caveat in preflight result, got %#v", got)
	}
}

func TestStationAIValidateAnswer(t *testing.T) {
	valid, issue := stationAIValidateAnswer(`{"status":"CONSTRAINT_VIOLATION"}`, stationAIIntentTrading)
	if valid {
		t.Fatalf("expected diagnostic answer to be rejected")
	}
	if issue == "" {
		t.Fatalf("expected rejection issue for diagnostic answer")
	}

	valid, _ = stationAIValidateAnswer("Keep filters stable and re-run later.", stationAIIntentTrading)
	if valid {
		t.Fatalf("expected trading answer without numbers to be rejected")
	}

	valid, issue = stationAIValidateAnswer(
		"Recommendation: raise min_daily_volume to 20 and min_item_profit to 1000000 ISK based on current rows.",
		stationAIIntentTrading,
	)
	if !valid {
		t.Fatalf("expected numeric trading answer to pass validation, issue=%q", issue)
	}
}

func TestNormalizeStationAIChatRequestStripsRuntimeContext(t *testing.T) {
	req := stationAIChatRequestPayload{
		Provider:    "openrouter",
		APIKey:      "test",
		Model:       "test-model",
		UserMessage: "hello",
		Context: stationAIContextPayload{
			Runtime: &stationAIRuntimeContext{
				Available: true,
			},
		},
	}
	_, _, _, validationErr := normalizeStationAIChatRequest(&req)
	if validationErr != "" {
		t.Fatalf("unexpected validation error: %s", validationErr)
	}
	if req.Context.Runtime != nil {
		t.Fatalf("expected normalizeStationAIChatRequest to clear client runtime context")
	}
}
