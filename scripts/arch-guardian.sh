#!/bin/bash
# arch-guardian.sh
#
# Watches ./internal/ for file changes via archlint watch.
# Runs two checks on every change:
#   1. archlint scan   - layer violations (structural)
#   2. archlint callgraph - behavioral cycles (if callgraph-entries.yaml exists)
# On violation: writes to VIOLATIONS_FILE.
# On fix detected: clears VIOLATIONS_FILE.
# Claude Code reads the file and fixes violations.

ARCHLINT="${ARCHLINT:-../archlint-repo/bin/archlint}"
CONFIG=".archlint.yaml"
DIR="./internal/"
WATCH_LOG="/tmp/arch-guardian-watch.log"
export VIOLATIONS_FILE="${VIOLATIONS_FILE:-/tmp/arch-violations.txt}"

cleanup() {
    rm -f "$WATCH_LOG"
    kill "$WATCH_PID" 2>/dev/null
    echo ""
    echo "Guardian stopped."
    exit 0
}
trap cleanup INT TERM

rm -f "$WATCH_LOG"
> "$VIOLATIONS_FILE"

$ARCHLINT watch "$DIR" >> "$WATCH_LOG" 2>&1 &
WATCH_PID=$!

echo "=== Architecture Guardian ==="
echo "Watching:        $DIR"
echo "Config:          $CONFIG"
echo "Violations file: $VIOLATIONS_FILE"
echo ""

LAST_SIZE=0

while true; do
    CURRENT_SIZE=$(wc -c < "$WATCH_LOG" 2>/dev/null | tr -d ' ')
    CURRENT_SIZE=${CURRENT_SIZE:-0}

    if [ "$CURRENT_SIZE" -gt "$LAST_SIZE" ]; then
        LAST_SIZE=$CURRENT_SIZE
        ts=$(date '+%H:%M:%S')
        violations=""

        # Check 1: layer violations (structural)
        scan_result=$($ARCHLINT scan "$DIR" --config "$CONFIG" 2>&1) || true
        if echo "$scan_result" | grep -q "FAILED"; then
            violations="$scan_result"
        fi

        # Check 2: behavioral cycles (only if callgraph-entries.yaml exists)
        if [ -f "callgraph-entries.yaml" ] && [ -z "$violations" ]; then
            while IFS= read -r entry; do
                [[ "$entry" =~ ^#.*$ || -z "$entry" ]] && continue
                $ARCHLINT callgraph "$DIR" --entry "$entry" --no-puml > /dev/null 2>&1 || true
                if [ -f "callgraphs/callgraph.yaml" ] && grep -q "cycles_detected: [1-9]" callgraphs/callgraph.yaml; then
                    violations=$(cat callgraphs/callgraph.yaml)
                    break
                fi
            done < <(grep -v '^#' callgraph-entries.yaml | grep -v '^entries:' | sed 's/^[[:space:]]*-[[:space:]]*//')
        fi

        if [ -n "$violations" ]; then
            echo "[$ts] VIOLATION -> $VIOLATIONS_FILE"
            echo "$violations" > "$VIOLATIONS_FILE"
        else
            echo "[$ts] OK"
            > "$VIOLATIONS_FILE"
        fi
    fi

    sleep 1
done
