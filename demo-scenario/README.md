# Live Demo Scenario: Stachka 2026

Realistic live demo for the archlint talk at Stachka 2026 conference.

Total runtime: ~3 minutes.

The story: a developer builds an order HTTP handler step by step.
Each feature addition looks reasonable. archlint watches in real time
and catches the moment fan-out crosses the limit.

---

## Setup (before the talk, ~2 min)

### 1. Open two terminals side by side

Terminal A (editor/copy-paste), Terminal B (archlint watch output).

### 2. Create the working file

```bash
# Create the file that the audience will watch change
cp demo-scenario/step0-clean.go internal/handler/order.go
```

### 3. Start archlint in watch mode (Terminal B)

```bash
archlint watch ./internal/handler/ --max-fan-out 5
```

Expected output:
```
watching ./internal/handler/ ...
OK  handler/order.go  fan-out=3
```

The terminal stays open and re-runs on every file save.

### 4. Open order.go in your editor (Terminal A)

Audience should see both terminals on screen.

---

## Step 0 - Clean handler (0:00)

**What to do:** Show the current state of `internal/handler/order.go`.

**What the audience sees:**
```
OK  handler/order.go  fan-out=3
```

**What to say (30 sec):**

> "Here is a fresh OrderHandler. It has three imports:
> the domain model, the service layer, and net/http.
> Three dependencies - clean, focused, easy to test.
> archlint is watching in the background. Fan-out is 3, limit is 5.
> Let's add some features and see what happens."

---

## Step 1 - Add logging (0:30)

**What to do:**
```bash
cp demo-scenario/step1-logging.go internal/handler/order.go
```

**What the audience sees:**
```
OK  handler/order.go  fan-out=5
```

**What to say (30 sec):**

> "First thing: add structured logging. Reasonable.
> We import fmt and log - two more packages.
> Fan-out is now 5. Still green, exactly at the limit.
> This is a developer making a sensible decision."

---

## Step 2 - Add caching (1:00)

**What to do:**
```bash
cp demo-scenario/step2-caching.go internal/handler/order.go
```

**What the audience sees:**
```
VIOLATION  handler/order.go  fan-out=7  (limit: 5)
```

**What to say (40 sec):**

> "Performance ticket arrives: 'GET /orders is slow, add caching.'
> Developer adds a map with a mutex and TTL. Makes sense, right?
> Two new imports: sync and time.
>
> And there it is - the violation fires immediately.
> Fan-out is 7. The handler is doing three things now:
> HTTP, logging, AND cache management.
> This is not a style nitpick. Each extra dependency
> is another reason for this file to change."

---

## Step 3 - Add metrics (1:40)

**What to do:**
```bash
cp demo-scenario/step3-metrics.go internal/handler/order.go
```

**What the audience sees:**
```
VIOLATION  handler/order.go  fan-out=8  (limit: 5)
```

**What to say (30 sec):**

> "Monitoring team asks for request counters.
> One more import - expvar. Now fan-out is 8.
> Each addition looks justified in isolation.
> This is how god files are born: not by malice,
> but by three reasonable decisions in a row.
> Without a linter watching, this lands in code review
> two weeks and five PRs later."

---

## Step 4 - Fix with facades (2:10)

**What to do:**
```bash
cp demo-scenario/step4-fixed.go internal/handler/order.go
```

**What the audience sees:**
```
OK  handler/order.go  fan-out=4
```

**What to say (40 sec):**

> "The fix: extract OrderCache and OrderMetrics as interfaces.
> The handler declares what it needs - two narrow contracts.
> It does not know about sync, time, or expvar anymore.
> Those move to separate implementation files - one responsibility each.
>
> Handler back to 4 imports. Green.
> The implementations can change independently.
> You can swap Redis for in-memory cache without touching the handler.
> You can add Prometheus without touching the handler.
>
> This is what archlint enforces: not a style rule,
> but a structural property of the codebase."

---

## Close (2:50)

**What to say (10 sec):**

> "Three minutes, one real developer workflow, one violation caught live.
> Put this in CI and the violation never reaches code review."

---

## Troubleshooting

### archlint not found

```bash
# Build from source
cd ~/projects/archlint-repo
go build -o bin/archlint ./cmd/archlint
export PATH=$PATH:$(pwd)/bin
```

### Watch mode not available

Use `archlint analyze` and re-run manually after each step:

```bash
archlint analyze ./internal/handler/ --max-fan-out 5
```

### File not updating

Make sure you are copying to the correct path:
```bash
ls -la internal/handler/order.go
```

---

## File reference

| File | Fan-out | Status |
|------|---------|--------|
| step0-clean.go | 3 | OK |
| step1-logging.go | 5 | OK (at limit) |
| step2-caching.go | 7 | VIOLATION |
| step3-metrics.go | 8 | VIOLATION (worse) |
| step4-fixed.go | 4 | OK (fixed) |

Import breakdown per step:

**step0:** model, service, net/http

**step1:** model, service, net/http, fmt, log

**step2:** model, service, net/http, fmt, log, sync, time

**step3:** model, service, net/http, fmt, log, sync, time, expvar

**step4:** model, service, net/http, log (cache and metrics behind interfaces)

---

## Why this scenario works for an audience

- The developer never does anything wrong on purpose
- Every addition (logging, caching, metrics) is a real production need
- The violation emerges naturally from accumulated reasonable decisions
- The fix is architectural, not cosmetic - audiences can see the difference
- Total steps fit in 3 minutes with room for questions
