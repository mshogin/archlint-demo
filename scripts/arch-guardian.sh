#!/bin/bash
# arch-guardian.sh
#
# Watches ./internal/ for file changes via archlint watch.
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
        result=$($ARCHLINT scan "$DIR" --config "$CONFIG" 2>&1) || true
        ts=$(date '+%H:%M:%S')

        if echo "$result" | grep -q "FAILED"; then
            echo "[$ts] VIOLATION -> $VIOLATIONS_FILE"
            echo "$result" > "$VIOLATIONS_FILE"
        else
            echo "[$ts] OK"
            > "$VIOLATIONS_FILE"
        fi
    fi

    sleep 1
done
