# archlint demo - Makefile for live demo on stage
# Usage: make step0, make step1, make step2, make step3, make step4

ARCHLINT ?= ../archlint-repo/bin/archlint
CONFIG = --config .archlint.yaml

SHELL := /bin/bash -x

# Remove //go:build ignore and copy to target
define copy_step
	sed '1d' demo-scenario/$(1) > $(2)
endef

## Step 0: Clean state (no violations, no cycles)
step0:
	$(call copy_step,step0-clean.go,internal/handler/order.go)
	$(call copy_step,step0-behavior-clean.go,internal/service/inventory_service.go)
	$(ARCHLINT) scan ./internal/ $(CONFIG)

## Step 1: Introduce layer violation (handler imports repo, with comments)
step1:
	$(call copy_step,step1-quick-fix.go,internal/handler/order.go)
	$(ARCHLINT) scan ./internal/ $(CONFIG)

## Step 1.1: Introduce layer violation + handler test (clean diff, good for demos)
step1.1:
	$(call copy_step,step1.1-violation.go,internal/handler/order.go)
	cp demo-scenario/step1.1-handler-test.go tests/handler_cache_test.go
	$(ARCHLINT) scan ./internal/ $(CONFIG)

## Step 2.1: Introduce behavioral cycle (guardian-detectable via callgraph)
step2.1:
	$(call copy_step,step2.1-fulfillment-cycle.go,internal/service/fulfillment.go)
	cp demo-scenario/step2.1-fulfillment-test.go tests/fulfillment_test.go
	cp demo-scenario/step2.1-callgraph-entries.yaml callgraph-entries.yaml
	$(ARCHLINT) callgraph ./internal --entry "internal/service.FulfillmentService.FulfillOrder" --no-puml

## Step 2: Introduce behavioral cycle
step2:
	$(call copy_step,step1-behavior-cycle.go,internal/service/inventory_service.go)
	$(ARCHLINT) callgraph ./internal --entry "internal/service.OrderServiceWithCycle.CreateOrderWithCycle" --no-puml

## Step 3: Fix layer violation
step3:
	$(call copy_step,step3-fixed.go,internal/handler/order.go)
	$(ARCHLINT) scan ./internal/ $(CONFIG)

## Step 4: Fix behavioral cycle
step4:
	$(call copy_step,step0-behavior-clean.go,internal/service/inventory_service.go)
	$(ARCHLINT) callgraph ./internal --entry "internal/service.OrderService.CreateOrder" --no-puml

## Collect full architecture graph
collect:
	$(ARCHLINT) collect .

## Watch mode (Ctrl+C to stop)
watch:
	$(ARCHLINT) watch ./internal/

## Run tests
test:
	go test ./tests/...

## Architecture Guardian - watch + auto-fix violations via Claude (Ctrl+C to stop)
guardian:
	@bash scripts/arch-guardian.sh

## Full demo sequence
demo: step0 step1 step3 step2 step4
	@echo ""
	@echo "=== DEMO COMPLETE ==="

## Reset to clean state
reset: step0
	rm -f tests/handler_cache_test.go tests/fulfillment_test.go
	$(call copy_step,step0-fulfillment-clean.go,internal/service/fulfillment.go)
	rm -f callgraph-entries.yaml

.PHONY: step0 step1 step1.1 step2 step2.1 step3 step4 collect watch test guardian demo reset
