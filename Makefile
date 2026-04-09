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

## Step 1: Introduce layer violation (handler imports repo)
step1:
	$(call copy_step,step1-quick-fix.go,internal/handler/order.go)
	$(ARCHLINT) scan ./internal/ $(CONFIG)

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

## Full demo sequence
demo: step0 step1 step3 step2 step4
	@echo ""
	@echo "=== DEMO COMPLETE ==="

## Reset to clean state
reset: step0

.PHONY: step0 step1 step2 step3 step4 collect watch demo reset
