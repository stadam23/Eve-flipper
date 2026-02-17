# Math Audit Phase 1 (Academic Baseline)

Date: 2026-02-17
Scope order: Flipper -> Regional Arbitrage -> Route Planner -> Contract Arbitrage
Next phase: Station Trading (after phase 1 closure)

## Goal
Build a strict mathematical baseline for trade/economic calculations so every core module satisfies explicit invariants and deterministic testability.

## Core principles
- Unit consistency: ISK, ISK/unit, %, units/day, ISK/day are never mixed.
- Execution-aware economics: ranking/filtering must prefer realizable fills over top-of-book illusion.
- Conservatism under uncertainty: when data quality is low, outputs must degrade predictably.
- Monotonicity: stricter filters never increase feasible set unexpectedly.
- Numerical safety: no NaN/Inf in persisted or API-facing metrics.

## Global invariants
- Non-negativity:
  - volume-like metrics >= 0
  - jumps >= 0
  - fill probability in [0, 1]
- Mass balance:
  - S2B + BfS = DailyVolume (within floating tolerance)
- Boundedness:
  - ratios finite; division-by-zero paths return bounded defaults
- Execution feasibility:
  - if CanFill=false for desired quantity, reported executable quantity must be reduced or dropped
- Profit coherence:
  - TotalProfit ~= ProfitPerUnit * Units (when same execution model)
  - ProfitPerJump ~= TotalProfit / TotalJumps when TotalJumps > 0

## Module-specific audit targets

### 1) Flipper
- Validate depth-aware expected profit and real margin.
- Validate S2B/BfS split monotonicity against order-book imbalance.
- Validate history-dependent filters: active history filters must exclude rows without history.

### 2) Regional Arbitrage
- Reuse Flipper invariants under multi-region context.
- Validate cross-region metrics remain unit-consistent and finite.
- Validate ranking stability under equal-profit ties.

### 3) Route Planner
- Validate each hop profit > 0 and jumps in (0, MaxTradeJumps].
- Validate route aggregates:
  - TotalProfit = sum hop profits
  - TotalJumps = sum hop jumps
  - ProfitPerJump = TotalProfit / TotalJumps
- Validate no system revisit within a route.

### 4) Contract Arbitrage
- Validate fill model monotonicity:
  - estimateFillDays increases with quantity and decreases with daily volume.
  - fillProbabilityWithinDays increases with horizon and decreases with fillDays.
- Validate probability bounds [0, 1] including Inf/zero edge cases.
- Validate conservative expected profit path does not exceed unconstrained profit when risk penalties apply.

## Definition of done for phase 1
- Automated invariant tests exist and pass in CI (`go test ./...`).
- No invariant violations under deterministic edge-case suites.
- Findings and risk notes are documented and reproducible.

## Out of scope for phase 1
- Full calibration/re-fitting of station-trading scores (CTS/OBDS/SDS) and advanced microstructure models.
- Major algorithmic redesign (beam/objective rewrite) without benchmark harness.