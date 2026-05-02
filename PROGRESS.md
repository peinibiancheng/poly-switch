# Project Progress: `poly-switch` TUI Optimization

## Current Task
**Optimize TUI design to match dashboard-style layout.**

## Status: ✅ Layout Implementation Complete

### Completed
- [x] Created detailed implementation plan in `/home/dev/.claude/plans/moonlit-popping-boole.md`
- [x] Analyzed current TUI implementation in `internal/ui/model.go`
- [x] Investigated data structures in `internal/core/app.go` and `internal/ui/model.go`
- [x] **Style Definition**: Added new `lipgloss` styles for dashboard (header, labels, status indicators, separators)
- [x] **Layout Implementation**: Refactored `langListView` to two-column layout (Left: Language List | Right: Environment Details)
- [x] **Env Details Panel**: Shows Status (● Active/Inactive), Version, Binary Path, Home Paths count, Detected versions count
- [x] **Consistent Header**: All views now share a unified header with purple background (`#1E1B4B`)
- [x] **PATH Warning**: Properly positioned below the two-column layout
- [x] **Build Verification**: `go build` and `go vet` pass clean

### Key Changes (internal/ui/model.go)
- **New styles**: `headerStyle`, `detailTitleStyle`, `labelStyle`, `valueStyle`, `statusActiveStyle`, `statusInactiveStyle`, `separatorStyle`
- **New method**: `envDetailsView()` — renders right-column environment details for the currently selected language
- **Updated method**: `langListView()` — now uses `lipgloss.JoinHorizontal` to place language list (40% width) and details panel side by side
- **Updated methods**: `loadingView()`, `versionListView()`, `applyingView()`, `doneView()` — all use consistent `headerStyle`
- **Window sizing**: Language list width set to 40% of total width (min 30 chars)

### Remaining / Future
- [ ] Run TUI and verify visual alignment
- [ ] Potentially enhance `versionListView` with a similar details panel
- [ ] Add `viewedLanguages` tracking or highlight animation on selection
