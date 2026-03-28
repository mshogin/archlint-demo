# Live Demo Scenario: Stachka 2026

Realistic live demo for the archlint talk at Stachka 2026 conference.

Total runtime: ~3 minutes.

The story: a developer "quickly fixes" a performance bug by importing the repo
directly from the handler, bypassing the service layer. This is the most
common real-world architecture violation - every developer has done it under
deadline pressure. archlint catches it the moment the file is saved.

---

## Setup (before the talk, ~2 min)

### 1. Open two terminals side by side

Terminal A (editor/copy-paste), Terminal B (archlint watch output).

### 2. Create the working file

```bash
cp demo-scenario/step0-clean.go internal/handler/order.go
```

### 3. Start archlint in watch mode (Terminal B)

```bash
archlint watch ./internal/handler/ --layers handler:1,service:2,repo:3
```

Expected output:
```
watching ./internal/handler/ ...
OK  handler/order.go  violations=0
```

The terminal stays open and re-runs on every file save.

### 4. Open order.go in your editor (Terminal A)

Audience should see both terminals on screen.

---

## Step 0 - Clean handler (0:00)

**What to do:** Show the current state of `internal/handler/order.go`.

**What the audience sees:**
```
OK  handler/order.go  violations=0
```

**What to say (30 sec):**

> "Here is a fresh OrderHandler. Three imports:
> net/http, model, and the service layer.
> The handler only talks to the service. Service talks to repo.
> One direction. No shortcuts.
> archlint is watching in the background. Zero violations.
> Now let's simulate a real Monday morning."

---

## Step 1 - "Quick fix" (0:30)

**What to do:**
```bash
cp demo-scenario/step1-quick-fix.go internal/handler/order.go
```

**What the audience sees:**
```
VIOLATION  handler/order.go
  handler -> repo  (forbidden: must go through service layer)
  1 violation
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

## Step 2 - Making it worse (1:15)

**What to do:**
```bash
cp demo-scenario/step2-worse.go internal/handler/order.go
```

**What the audience sees:**
```
VIOLATION  handler/order.go
  handler -> repo  (forbidden: must go through service layer)
  2 violations
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

## Step 3 - Fixed (1:45)

**What to do:**
```bash
cp demo-scenario/step3-fixed.go internal/handler/order.go
```

**What the audience sees:**
```
OK  handler/order.go  violations=0
```

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

## Close (2:45)

**What to say (15 sec):**

> "Three minutes. One violation pattern that every team ships at least once.
> Put archlint in CI and the 'quick fix' gets a review comment before merge,
> not six months later when the handler has four repo imports and nobody
> remembers why."

---

## Troubleshooting

### archlint not found

```bash
cd ~/projects/archlint-repo
go build -o bin/archlint ./cmd/archlint
export PATH=$PATH:$(pwd)/bin
```

### Watch mode not available

Use `archlint analyze` and re-run manually after each step:

```bash
archlint analyze ./internal/handler/
```

### File not updating

Make sure you are copying to the correct path:
```bash
ls -la internal/handler/order.go
```

---

## File reference

| File | Violations | Status |
|------|------------|--------|
| step0-clean.go | 0 | OK |
| step1-quick-fix.go | 1 | VIOLATION (handler -> repo) |
| step2-worse.go | 2 | VIOLATION (handler -> repo x2) |
| step3-fixed.go | 0 | OK (fixed) |

Import breakdown per step:

**step0:** net/http, model, service

**step1:** net/http, model, service, repo (VIOLATION: +repo direct)

**step2:** net/http, model, service, repo (VIOLATION: +OrderCache +UserRepo from repo)

**step3:** net/http, model, service (repo hidden behind service interface)

---

## Why this scenario works for an audience

- The developer's reasoning is completely sound under deadline pressure
- "Import repo directly - it's faster" is something everyone has thought
- The violation is structural, not stylistic - easy to explain why it matters
- The fix is real and fast - no magic, just moving code to the right layer
- The archlint output names the exact forbidden dependency
- Total steps fit in 3 minutes with room for questions
