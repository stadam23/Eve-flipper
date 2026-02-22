package api

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"
)

type stationAIWikiEvalCase struct {
	Name      string
	Query     string
	Intent    stationAIIntentKind
	WantPages []string
}

type stationAIWikiEvalResult struct {
	CaseName string
	TopPages []string
	Hit      bool
}

type stationAIWikiPageDoc struct {
	Slug    string
	Content string
}

func TestStationAIWikiRetrievalEvalHarness(t *testing.T) {
	idx := buildStationAIWikiEvalIndex(t)
	cases := []stationAIWikiEvalCase{
		{
			Name:      "abbrev_cts",
			Query:     "what is cts",
			Intent:    stationAIIntentTrading,
			WantPages: []string{"Station-Trading"},
		},
		{
			Name:      "abbrev_sds",
			Query:     "what is sds",
			Intent:    stationAIIntentTrading,
			WantPages: []string{"Station-Trading"},
		},
		{
			Name:      "abbrev_pvi",
			Query:     "what is pvi",
			Intent:    stationAIIntentTrading,
			WantPages: []string{"Station-Trading"},
		},
		{
			Name:      "abbrev_bvs",
			Query:     "what is bvs",
			Intent:    stationAIIntentTrading,
			WantPages: []string{"Station-Trading"},
		},
		{
			Name:      "radius_structures",
			Query:     "radius scan structures include",
			Intent:    stationAIIntentTrading,
			WantPages: []string{"Radius-Scan"},
		},
		{
			Name:      "radius_how_it_works",
			Query:     "how radius scan works",
			Intent:    stationAIIntentProduct,
			WantPages: []string{"Radius-Scan"},
		},
		{
			Name:      "execution_done_ignored",
			Query:     "done ignored command center",
			Intent:    stationAIIntentTrading,
			WantPages: []string{"Execution-Plan"},
		},
		{
			Name:      "api_chat_endpoint",
			Query:     "station ai chat endpoint",
			Intent:    stationAIIntentDebug,
			WantPages: []string{"API-Reference"},
		},
		{
			Name:      "plex_omega",
			Query:     "omega plex usd spread",
			Intent:    stationAIIntentProduct,
			WantPages: []string{"PLEX-Dashboard"},
		},
		{
			Name:      "war_hotzones",
			Query:     "war tracker hot zones",
			Intent:    stationAIIntentProduct,
			WantPages: []string{"War-Tracker"},
		},
		{
			Name:      "execution_plan_rows",
			Query:     "build execution plan from rows",
			Intent:    stationAIIntentTrading,
			WantPages: []string{"Execution-Plan"},
		},
	}

	const k = 1
	ragRecall, ragResults := runStationAIWikiEvalCases(cases, func(tc stationAIWikiEvalCase) []string {
		ranked, _ := idx.hybridSearch(
			context.Background(),
			defaultStationAIWikiRepo,
			"en",
			tc.Query,
			tc.Intent,
			stationAIWikiTopK,
		)
		return stationAIUniqueTopPages(idx, ranked, k)
	})

	fallbackRecall, fallbackResults := runStationAIWikiEvalCases(cases, func(tc stationAIWikiEvalCase) []string {
		return stationAIFallbackWikiPages(idx, tc.Query, k)
	})

	t.Logf("station-ai wiki retrieval Recall@%d: rag=%.2f fallback=%.2f", k, ragRecall, fallbackRecall)
	for i := 0; i < len(cases); i++ {
		tc := cases[i]
		rag := ragResults[i]
		fallback := fallbackResults[i]
		t.Logf(
			"A/B %s: rag_hit=%v rag_top=%v | fallback_hit=%v fallback_top=%v | want=%v",
			tc.Name,
			rag.Hit,
			rag.TopPages,
			fallback.Hit,
			fallback.TopPages,
			tc.WantPages,
		)
	}

	if ragRecall < 0.80 {
		t.Fatalf("rag recall too low: got %.2f, want >= 0.80", ragRecall)
	}
	if ragRecall+0.05 < fallbackRecall {
		t.Fatalf("rag recall regressed vs fallback: rag=%.2f fallback=%.2f", ragRecall, fallbackRecall)
	}
}

func TestStationAIQueryHintPageSet(t *testing.T) {
	tests := []struct {
		name string
		q    string
		want string
	}{
		{name: "radius", q: "how radius scan works", want: "Radius-Scan"},
		{name: "cts", q: "what is cts and sds", want: "Station-Trading"},
		{name: "api", q: "station ai chat api endpoint", want: "API-Reference"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hints := stationAIQueryHintPageSet(tc.q)
			if len(hints) == 0 {
				t.Fatalf("expected non-empty hints for query %q", tc.q)
			}
			if _, ok := hints[tc.want]; !ok {
				t.Fatalf("expected hint for %q in query %q, got %+v", tc.want, tc.q, hints)
			}
		})
	}
}

func runStationAIWikiEvalCases(
	cases []stationAIWikiEvalCase,
	retriever func(stationAIWikiEvalCase) []string,
) (float64, []stationAIWikiEvalResult) {
	if len(cases) == 0 {
		return 0, nil
	}
	results := make([]stationAIWikiEvalResult, 0, len(cases))
	hits := 0
	for _, tc := range cases {
		top := retriever(tc)
		hit := stationAIContainsAnyPage(top, tc.WantPages)
		if hit {
			hits++
		}
		results = append(results, stationAIWikiEvalResult{
			CaseName: tc.Name,
			TopPages: top,
			Hit:      hit,
		})
	}
	return float64(hits) / float64(len(cases)), results
}

func stationAIContainsAnyPage(got []string, want []string) bool {
	if len(got) == 0 || len(want) == 0 {
		return false
	}
	wantSet := make(map[string]struct{}, len(want))
	for _, w := range want {
		wantSet[strings.TrimSpace(w)] = struct{}{}
	}
	for _, g := range got {
		if _, ok := wantSet[strings.TrimSpace(g)]; ok {
			return true
		}
	}
	return false
}

func stationAIUniqueTopPages(idx *stationAIWikiRAGIndex, ranked []stationAIWikiRankedDoc, k int) []string {
	if k <= 0 {
		return nil
	}
	seen := make(map[string]struct{}, k)
	out := make([]string, 0, k)
	for _, doc := range ranked {
		if doc.Index < 0 || doc.Index >= len(idx.Chunks) {
			continue
		}
		slug := strings.TrimSpace(idx.Chunks[doc.Index].PageSlug)
		if slug == "" {
			continue
		}
		if _, ok := seen[slug]; ok {
			continue
		}
		seen[slug] = struct{}{}
		out = append(out, slug)
		if len(out) >= k {
			break
		}
	}
	return out
}

func stationAIFallbackWikiPages(idx *stationAIWikiRAGIndex, query string, k int) []string {
	if k <= 0 || idx == nil || len(idx.Chunks) == 0 {
		return nil
	}
	type pageCandidate struct {
		Slug  string
		Score int
		Order int
	}

	pageOrder := make([]string, 0, 16)
	pageDocs := make(map[string]*strings.Builder, 16)
	for _, ch := range idx.Chunks {
		slug := strings.TrimSpace(ch.PageSlug)
		if slug == "" {
			continue
		}
		if _, ok := pageDocs[slug]; !ok {
			pageDocs[slug] = &strings.Builder{}
			pageOrder = append(pageOrder, slug)
		}
		doc := pageDocs[slug]
		doc.WriteString(ch.Section)
		doc.WriteByte('\n')
		doc.WriteString(ch.Content)
		doc.WriteByte('\n')
	}
	terms := aiKeywordTerms(query)
	candidates := make([]pageCandidate, 0, len(pageDocs))
	for i, slug := range pageOrder {
		doc := pageDocs[slug].String()
		score, _ := aiScoreAndSnippet(doc, terms)
		if score <= 0 {
			continue
		}
		candidates = append(candidates, pageCandidate{Slug: slug, Score: score, Order: i})
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].Order < candidates[j].Order
		}
		return candidates[i].Score > candidates[j].Score
	})
	out := make([]string, 0, k)
	if len(candidates) == 0 {
		for _, slug := range pageOrder {
			out = append(out, slug)
			if len(out) >= k {
				break
			}
		}
		return out
	}
	for _, c := range candidates {
		out = append(out, c.Slug)
		if len(out) >= k {
			break
		}
	}
	return out
}

func buildStationAIWikiEvalIndex(t *testing.T) *stationAIWikiRAGIndex {
	t.Helper()
	docs := []stationAIWikiPageDoc{
		{
			Slug: "Home",
			Content: `# Home
EVE Flipper wiki home page.
Overview of station trading, route trading, and dashboard tabs.
`,
		},
		{
			Slug: "Station-Trading",
			Content: `# Station Trading
## Metrics
Composite trade score estimates expected upside per item.
Scam detection score estimates how likely fake spread patterns are.
Price volatility index tracks unstable price movement.
Buy sell flow ratio helps detect balance between buy and sell pressure.
## Advanced Filters
Minimum daily volume, minimum item profit, and period ROI set scan strictness.
`,
		},
		{
			Slug: "Radius-Scan",
			Content: `# Radius Scan
Radius scan expands station trading beyond one station by jump distance.
You can include structures and limit by region/system scope.
`,
		},
		{
			Slug: "Execution-Plan",
			Content: `# Execution Plan
Execution plan builds action rows from scan output.
Command center supports done and ignored states for each row.
`,
		},
		{
			Slug: "API-Reference",
			Content: `# API Reference
Station AI chat endpoint: POST /api/auth/station/ai/chat.
Station scan endpoint: POST /api/scan/station.
`,
		},
		{
			Slug: "PLEX-Dashboard",
			Content: `# PLEX Dashboard
The dashboard tracks NES price, omega USD parity, and spread conditions.
`,
		},
		{
			Slug: "War-Tracker",
			Content: `# War Tracker
War tracker identifies hot zones and conflict-driven demand spikes.
`,
		},
	}

	chunks := make([]stationAIWikiChunk, 0, 64)
	embedInputs := make([]string, 0, 64)
	seq := 0
	for _, doc := range docs {
		page := strings.ReplaceAll(doc.Slug, "-", " ")
		sections := stationAIExtractMarkdownSections(doc.Content)
		if len(sections) == 0 {
			sections = []stationAIWikiSection{{Content: strings.TrimSpace(doc.Content)}}
		}
		for _, sec := range sections {
			sectionText := strings.TrimSpace(sec.Content)
			if sectionText == "" {
				continue
			}
			sectionChunks := stationAIRecursiveSplitWithOverlap(sectionText, 240, 40)
			sectionLabel := page
			if len(sec.Headings) > 0 {
				sectionLabel = strings.Join(sec.Headings, " > ")
			}
			breadcrumb := stationAIBuildBreadcrumb(page, sec.Headings)
			for _, chunkText := range sectionChunks {
				chunkText = strings.TrimSpace(chunkText)
				if chunkText == "" {
					continue
				}
				seq++
				ch := stationAIWikiChunk{
					ID:         fmt.Sprintf("%s:%d", doc.Slug, seq),
					SourcePath: doc.Slug + ".md",
					Page:       page,
					PageSlug:   doc.Slug,
					Section:    sectionLabel,
					Breadcrumb: breadcrumb,
					Locale:     "en",
					Content:    chunkText,
					TokenCount: stationAITokenCount(chunkText),
				}
				chunks = append(chunks, ch)
				embedInputs = append(embedInputs, breadcrumb+"\n"+chunkText)
			}
		}
	}
	if len(chunks) == 0 {
		t.Fatal("eval index build produced no chunks")
	}

	vectors := stationAIEmbedLocal(embedInputs, stationAIWikiRAGLocalEmbeddingDim)
	if len(vectors) != len(chunks) {
		t.Fatalf("embedding count mismatch: vectors=%d chunks=%d", len(vectors), len(chunks))
	}
	idx := &stationAIWikiRAGIndex{
		Repo:           defaultStationAIWikiRepo,
		BuiltAt:        time.Now().UTC().Format(time.RFC3339),
		FileHashes:     map[string]string{},
		EmbeddingKind:  "local",
		EmbeddingModel: "local-hash",
		EmbeddingDim:   stationAIWikiRAGLocalEmbeddingDim,
		Chunks:         chunks,
		Vectors:        vectors,
	}
	idx.rebuildLexical()
	return idx
}
