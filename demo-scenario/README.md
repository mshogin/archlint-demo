# Live Demo Scenario: Stachka 2026

Realistic live demo for the archlint talk at Stachka 2026 conference.

Total runtime: ~3 minutes (structural demo) + ~2 minutes (behavioral demo, optional).

The story: a developer "quickly fixes" a performance bug by importing the repo
directly from the handler, bypassing the service layer. This is the most
common real-world architecture violation - every developer has done it under
deadline pressure. archlint catches it the moment the file is saved.

---

## Quick Reference

| Mode | Command | Use case |
|------|---------|----------|
| Manual | `cp ... && archlint scan` | Slides, controlled pacing |
| Watch | `archlint watch` + `cp ...` | Live demo, reactive output |

---

## Mode A: Manual (step by step cp + scan)

Use this mode when you want full control over timing, e.g. during a slide deck
where you run commands after each talking point.

### Setup

```bash
# Build archlint (if not installed)
cd ~/projects/archlint-repo
go build -o bin/archlint ./cmd/archlint
export PATH=$PATH:$(pwd)/bin

# Go to the demo project
cd ~/projects/archlint-demo

# Reset to clean state
cp demo-scenario/step0-clean.go internal/handler/order.go
```

### Step 0 - Show the clean baseline

```bash
archlint scan ./internal/ --config .archlint.yaml
```

Expected output:
```
config: .archlint.yaml
PASSED: No violations found (threshold: 0)
```

### Step 1 - Inject the "quick fix" violation

```bash
cp demo-scenario/step1-quick-fix.go internal/handler/order.go
archlint scan ./internal/ --config .archlint.yaml
```

Expected output:
```
config: .archlint.yaml
FAILED: 1 violations found (threshold: 0)

[ERROR] [layer-violation] Forbidden dependency: internal/handler (handler) -> demo/internal/repo (repository)
  target: internal/handler
```

### Step 2 - Make it worse (second direct repo import)

```bash
cp demo-scenario/step2-worse.go internal/handler/order.go
archlint scan ./internal/ --config .archlint.yaml
```

Expected output (same violation, now two types in the forbidden import):
```
config: .archlint.yaml
FAILED: 1 violations found (threshold: 0)

[ERROR] [layer-violation] Forbidden dependency: internal/handler (handler) -> demo/internal/repo (repository)
  target: internal/handler
```

### Step 3 - Fix

```bash
cp demo-scenario/step3-fixed.go internal/handler/order.go
archlint scan ./internal/ --config .archlint.yaml
```

Expected output:
```
config: .archlint.yaml
PASSED: No violations found (threshold: 0)
```

---

## Mode B: Watch mode (archlint watch + cp files)

Use this mode for a live terminal demo where the linter reacts automatically.
Run two terminals side by side.

### Terminal B - Start the watcher

```bash
cd ~/projects/archlint-demo
archlint watch ./internal/
```

The watcher performs an immediate scan on start and then re-scans on every `.go`
file change. Leave this terminal visible to the audience.

Initial output (clean state):
```
[archlint] Watching ./internal/ ...
[archlint] Scanning...
[archlint] 0 errors, 18 warnings
[archlint] Last scan: HH:MM:SS
```

Note: the watch command runs all rules (including DIP, feature-envy, etc.).
Layer violations show as errors; SOLID warnings show as warnings.
The scan command with `--config` only checks the rules enabled in .archlint.yaml.

### Terminal A - Copy step files to inject/fix violations

```bash
# Step 0: reset clean baseline (already done above)
cp demo-scenario/step0-clean.go internal/handler/order.go

# Step 1: quick fix - handler imports repo directly
cp demo-scenario/step1-quick-fix.go internal/handler/order.go

# Step 2: making it worse - second direct repo import
cp demo-scenario/step2-worse.go internal/handler/order.go

# Step 3: fix - repo hidden behind service interface
cp demo-scenario/step3-fixed.go internal/handler/order.go
```

After each copy, Terminal B automatically re-scans and prints the result.

### Note about //go:build ignore

The demo-scenario step files all begin with `//go:build ignore` so they do not
compile as part of the `demo/internal/handler` package until copied over.
When `cp` overwrites `internal/handler/order.go` the build tag is gone and the
file becomes the active handler code.

If you need to cat a step file without the build tag line (e.g. to pipe it
somewhere), use:

```bash
sed '1d' demo-scenario/step1-quick-fix.go
```

This removes the first line (`//go:build ignore`) and leaves valid Go source.

---

## Structural Demo Walkthrough (3 min)

### Step 0 - Clean handler (0:00)

**What to do:** Show the current state or copy step0.

**What the audience sees:** `PASSED: No violations found`

**What to say (30 sec):**

> "Here is a fresh OrderHandler. Three imports:
> net/http, model, and the service layer.
> The handler only talks to the service. Service talks to repo.
> One direction. No shortcuts.
> archlint is watching in the background. Zero violations.
> Now let's simulate a real Monday morning."

---

### Step 1 - "Quick fix" (0:30)

**What to do:**
```bash
cp demo-scenario/step1-quick-fix.go internal/handler/order.go
```

**What the audience sees:**
```
FAILED: 1 violations found (threshold: 0)

[ERROR] [layer-violation] Forbidden dependency: internal/handler (handler) -> demo/internal/repo (repository)
```

**What to say (45 sec):**

> "Performance ticket arrives Friday at 4pm:
> 'GET /orders is slow, cache check is missing.'
>
> Developer thinks: I just need to check the cache before hitting the DB.
> The cache is in the repo package - I'll import it directly.
> It's faster than adding a new method to the service. Ship it.
>
> Watch Terminal B when I save the file.
>
> There it is. The violation fires immediately.
> handler -> repo. Forbidden dependency.
> The handler is now coupled to a concrete repo type.
> If you swap the cache implementation, you touch the handler.
> If you test the handler, you must wire a real cache."

---

### Step 2 - Making it worse (1:15)

**What to do:**
```bash
cp demo-scenario/step2-worse.go internal/handler/order.go
```

**What the audience sees:**
```
FAILED: 1 violations found

[ERROR] [layer-violation] Forbidden dependency: internal/handler (handler) -> demo/internal/repo (repository)
```

**What to say (30 sec):**

> "One week later. Same developer, same reasoning:
> 'I also need to invalidate the user's order count cache on create.
> That cache is in repo too. While I'm already importing repo - one more type.'
>
> Now two forbidden repo dependencies from the handler.
> Each addition looked justified in isolation.
> This is how layer shortcuts multiply.
> Without a linter, this lands in code review two weeks later
> buried in a 500-line PR."

---

### Step 3 - Fixed (1:45)

**What to do:**
```bash
cp demo-scenario/step3-fixed.go internal/handler/order.go
```

**What the audience sees:** `PASSED: No violations found`

**What to say (60 sec):**

> "The fix takes five minutes.
>
> Move the cache logic into the service layer.
> OrderService gets two methods: GetWithCache and Create.
> Cache check, DB fallback, cache populate - all inside service.
> Handler calls one method. It does not know a cache exists.
>
> Handler back to three imports. net/http, model, service.
> Zero violations.
>
> The service owns the OrderCache interface - a narrow contract.
> You can swap Redis for in-memory without touching the handler.
> You can test the handler with a mock service - no repo wiring needed.
>
> This is the point archlint enforces: not a style rule.
> A structural property. The boundary between layers stays a boundary."

---

### Close (2:45)

**What to say (15 sec):**

> "Three minutes. One violation pattern that every team ships at least once.
> Put archlint in CI and the 'quick fix' gets a review comment before merge,
> not six months later when the handler has four repo imports and nobody
> remembers why."

---

## Behavioral Graph Demo (optional, +2 min)

A second demo showing **behavioral cycles** in the call graph.
This is a different class of problem from import layer violations:
the cycle exists in the runtime call chain, not in the import graph.

### The story

Two services that "need each other": OrderService calls CheckInventory
to check stock. CheckInventory eventually calls GetOrderDetails to validate
order context before committing a reservation. GetOrderDetails calls back into
CheckInventory. Each call looks reasonable in isolation. Together they form a cycle.

In a monolith this is a hidden coupling. In a microservice split it becomes
a distributed deadlock. archlint callgraph detects it statically.

### Setup

No file changes needed. The cycle is already in `internal/service/order_service.go`:

- `CreateOrder` calls `CheckInventory`
- `GetOrderDetails` calls `CheckInventory`
- `CheckInventory` is resolved by archlint as external (the function is referenced but resolved at the module level)

To use the explicit cycle demo files from `demo-scenario/`, remove the `//go:build ignore` tag first:

```bash
# Show the clean version (no cycle)
sed '1d' demo-scenario/step0-behavior-clean.go > internal/service/order_service_clean.go

# Show the cycle version
sed '1d' demo-scenario/step1-behavior-cycle.go > /tmp/cycle_demo.go
# Inspect only - do not compile as part of the service package
```

### Run the behavioral demo

```bash
# Detect the behavioral call graph from the CreateOrder entry point
archlint callgraph ./internal --entry "internal/service.OrderService.CreateOrder" --no-puml
```

Expected output:
```
Analyzing code: ./internal (language: go)
Graph built: 5 nodes, 6 edges, depth 0
YAML: callgraphs/callgraph.yaml
```

The YAML at `callgraphs/callgraph.yaml` contains the full call chain.
Check `stats.cycles_detected` to see the cycle count.

### What the call chain looks like

```
CreateOrder
  -> CheckInventory               (inventory domain)
       -> ReserveForOrder         (inventory domain)
            -> GetOrderDetails    (order domain - crosses boundary)
                 -> CheckInventory  [CYCLE: already visited]
```

The cycle closes at `GetOrderDetails -> CheckInventory`.
Both domains are now coupled at runtime.

### What to say on stage (90 sec)

> "Layer violations are visible in the import graph.
> Behavioral cycles are invisible there - the imports look fine.
> Both OrderService and InventoryService are in the same package.
> No forbidden imports. archlint scan would not catch this.
>
> But archlint callgraph traces the actual call chain from an entry point.
> Watch what happens when I run it against CreateOrder.
>
> The YAML contains the call chain. Check cycles_detected.
>
> GetOrderDetails calls CheckInventory. That edge has cycle: true.
> The graph closed on itself.
>
> This means you cannot extract InventoryService into a separate microservice
> without also extracting its dependency on GetOrderDetails.
> Which means you cannot extract GetOrderDetails without CheckInventory.
> They are locked together at runtime.
>
> The static import graph looked clean. The behavioral graph is not."

---

## Troubleshooting

### archlint not found

```bash
cd ~/projects/archlint-repo
go build -o bin/archlint ./cmd/archlint
export PATH=$PATH:$(pwd)/bin
```

### Watch mode not available / watch hangs

Use `archlint scan` in manual mode and re-run after each step:

```bash
archlint scan ./internal/ --config .archlint.yaml
```

### File not updating / wrong output

Make sure you are copying to the correct path:
```bash
ls -la internal/handler/order.go
```

### Step file has //go:build ignore

The demo-scenario files include `//go:build ignore` to prevent them from
compiling as part of the package. `cp` removes the build tag by overwriting the
file. If you need to inspect a step file as plain Go source without the tag:

```bash
sed '1d' demo-scenario/step1-quick-fix.go
```

---

## File Reference

### Structural demo files

| File | Violations | Status |
|------|------------|--------|
| step0-clean.go | 0 | OK |
| step1-quick-fix.go | 1 | VIOLATION (handler -> repo) |
| step2-worse.go | 1 | VIOLATION (handler -> repo, 2 concrete types) |
| step3-fixed.go | 0 | OK (fixed) |

Import breakdown per step:

**step0:** net/http, model, service

**step1:** net/http, model, service, repo (VIOLATION: +repo direct, OrderCache)

**step2:** net/http, model, service, repo (VIOLATION: +OrderCache +UserRepo from repo)

**step3:** net/http, model, service (repo hidden behind service interface)

### Behavioral demo files

| File | Cycles | Description |
|------|--------|-------------|
| step0-behavior-clean.go | 0 | InventoryService is a leaf, no call back into order domain |
| step1-behavior-cycle.go | 1 | InventoryService calls GetOrderDetails, cycle closes |
| internal/service/order_service.go | structural | CreateOrder, GetOrderDetails with CheckInventory calls |
| internal/service/inventory_service.go | - | Clean version: InventoryServiceClean is a leaf |

---

## Why this scenario works for an audience

- The developer's reasoning is completely sound under deadline pressure
- "Import repo directly - it's faster" is something everyone has thought
- The violation is structural, not stylistic - easy to explain why it matters
- The fix is real and fast - no magic, just moving code to the right layer
- The archlint output names the exact forbidden dependency
- Total steps fit in 3 minutes with room for questions
