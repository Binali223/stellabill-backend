#!/bin/bash
# analyze_benchmarks.sh - Analyze benchmark results and detect regressions

set -e

if [ $# -lt 2 ]; then
    echo "Usage: $0 <baseline.txt> <new.txt> [threshold] [whitelist_comma_separated]"
    echo "Example: $0 baseline.txt new.txt 1.10 BenchmarkListPlans,BenchmarkListSubscriptions"
    exit 1
fi

BASELINE=$1
NEW=$2
THRESHOLD=${3:-1.20}  # Default 20% regression threshold
WHITELIST=${4:-""}

if [ ! -f "$BASELINE" ]; then
    echo "Error: Baseline file not found: $BASELINE"
    exit 1
fi

if [ ! -f "$NEW" ]; then
    echo "Error: New benchmark file not found: $NEW"
    exit 1
fi

echo "Analyzing benchmarks..."
echo "Baseline: $BASELINE"
echo "New: $NEW"
echo "Regression threshold: ${THRESHOLD}x"
if [ -n "$WHITELIST" ]; then
    echo "Whitelist: $WHITELIST"
fi
echo ""

# Check if benchstat is installed
if ! command -v benchstat &> /dev/null; then
    echo "Installing benchstat..."
    go install golang.org/x/perf/cmd/benchstat@latest
fi

# Run comparison
echo "=== Benchmark Comparison ==="
benchstat "$BASELINE" "$NEW" > comparison.txt || true
cat comparison.txt

echo ""
echo "=== Regression Analysis ==="

# Parse results and check for regressions
REGRESSIONS=0

# Helper to check if a benchmark is in the whitelist
is_whitelisted() {
    local bench_name=$1
    if [ -z "$WHITELIST" ]; then
        return 0 # No whitelist means all are whitelisted
    fi
    IFS=',' read -ra ADDR <<< "$WHITELIST"
    for i in "${ADDR[@]}"; do
        if [[ "$bench_name" == "$i"* ]]; then
            return 0
        fi
    done
    return 1
}

while IFS= read -r line; do
    # Skip header / blank lines
    [[ -z "$line" ]] && continue
    [[ "$line" =~ ^(name|goos|goarch|pkg|cpu|PASS|ok|---) ]] && continue
    
    # Look for lines with performance changes (e.g., +13.64%)
    if echo "$line" | grep -qE "\+[0-9]+\.[0-9]+%"; then
        # Extract benchmark name (first word)
        BENCH_NAME=$(echo "$line" | awk '{print $1}')
        
        # Check whitelist
        if ! is_whitelisted "$BENCH_NAME"; then
            echo "ℹ️  Skipping non-whitelisted benchmark: $BENCH_NAME"
            continue
        fi
        
        CHANGE=$(echo "$line" | grep -oE "\+[0-9]+\.[0-9]+" | head -1)
        PERCENT=$(echo "$CHANGE" | tr -d '+')
        
        # Convert to multiplier
        MULTIPLIER=$(echo "1 + $PERCENT / 100" | bc -l)
        
        # Check if exceeds threshold (using awk for floating point comparison)
        EXCEEDS=$(awk -v m="$MULTIPLIER" -v t="$THRESHOLD" 'BEGIN{print (m > t) ? 1 : 0}')
        if [ "$EXCEEDS" -eq 1 ]; then
            echo "⚠️  REGRESSION DETECTED: $line"
            REGRESSIONS=$((REGRESSIONS + 1))
        fi
    fi
done < comparison.txt

echo ""
if [ $REGRESSIONS -gt 0 ]; then
    echo "❌ Found $REGRESSIONS regression(s) exceeding ${THRESHOLD}x threshold"
    exit 1
else
    echo "✅ No significant regressions detected"
    exit 0
fi
