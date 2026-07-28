#!/bin/bash
# test_analyze_benchmarks.sh - Test the benchmark analysis script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
ANALYZE_SCRIPT="$SCRIPT_DIR/analyze_benchmarks.sh"
cd "$SCRIPT_DIR"

# Helper to create dummy benchstat output
create_benchstat_mock() {
    local file=$1
    shift
    echo "$@" > "$file"
}

echo "Running tests for analyze_benchmarks.sh..."

# 1. Missing baseline
if $ANALYZE_SCRIPT missing_baseline.txt new.txt 1.10 > output.log 2>&1; then
    echo "❌ Test Failed: Should exit 1 when baseline is missing"
    exit 1
fi
grep -q "Error: Baseline file not found" output.log || { echo "❌ Missing baseline error not found"; exit 1; }

# Create dummy baseline and new files to pass file checks
touch base_dummy.txt new_dummy.txt

# 2. Missing new
if $ANALYZE_SCRIPT base_dummy.txt missing_new.txt 1.10 > output.log 2>&1; then
    echo "❌ Test Failed: Should exit 1 when new file is missing"
    exit 1
fi
grep -q "Error: New benchmark file not found" output.log || { echo "❌ Missing new error not found"; exit 1; }

# Mock benchstat to just cat a provided comparison.txt to bypass actual benchstat execution
# Actually, the script runs `benchstat "$BASELINE" "$NEW" > comparison.txt`.
# If baseline and new aren't valid benchstat inputs, benchstat might fail.
# Let's create an executable mock of benchstat in the current directory and prepend it to PATH.
cat << 'EOF' > mock_benchstat
#!/bin/bash
if [ -f "mock_comparison.txt" ]; then
    cat mock_comparison.txt
fi
EOF
chmod +x mock_benchstat
export PATH="$SCRIPT_DIR:$PATH"
alias benchstat="mock_benchstat"
# Actually script runs command -v benchstat. So the PATH trick works.

# Test 3: No regressions
create_benchstat_mock mock_comparison.txt "
name                      old time/op  new time/op  delta
BenchmarkListPlans        10.0ms ± 2%  10.1ms ± 2%  +1.00%
"
if ! $ANALYZE_SCRIPT base_dummy.txt new_dummy.txt 1.10 > output.log 2>&1; then
    echo "❌ Test Failed: Should pass when no regressions > 10%"
    exit 1
fi

# Test 4: Regression detected > 10%
create_benchstat_mock mock_comparison.txt "
name                      old time/op  new time/op  delta
BenchmarkListPlans        10.0ms ± 2%  11.5ms ± 2%  +15.00%
"
if $ANALYZE_SCRIPT base_dummy.txt new_dummy.txt 1.10 > output.log 2>&1; then
    echo "❌ Test Failed: Should exit 1 when regression > 10%"
    exit 1
fi
grep -q "REGRESSION DETECTED" output.log || { echo "❌ Regression output missing"; exit 1; }

# Test 5: Regression detected but whitelisted out
create_benchstat_mock mock_comparison.txt "
name                      old time/op  new time/op  delta
BenchmarkListPlans        10.0ms ± 2%  11.5ms ± 2%  +15.00%
BenchmarkOther            10.0ms ± 2%  10.1ms ± 2%  +1.00%
"
if ! $ANALYZE_SCRIPT base_dummy.txt new_dummy.txt 1.10 "BenchmarkOther" > output.log 2>&1; then
    echo "❌ Test Failed: Should pass because failing benchmark is not whitelisted"
    cat output.log
    exit 1
fi
grep -q "Skipping non-whitelisted benchmark: BenchmarkListPlans" output.log || { echo "❌ Skipped output missing"; exit 1; }

# Test 6: Absent from baseline (New benchmark)
# Benchstat handles this by not showing a delta line with a percentage, or just showing empty.
# We will simulate a line without a percentage delta.
create_benchstat_mock mock_comparison.txt "
name                      old time/op  new time/op  delta
BenchmarkNewBenchmark     -            10.1ms ± 2%  -
"
if ! $ANALYZE_SCRIPT base_dummy.txt new_dummy.txt 1.10 > output.log 2>&1; then
    echo "❌ Test Failed: Should pass when benchmark is new (no delta)"
    exit 1
fi

rm -f base_dummy.txt new_dummy.txt mock_benchstat mock_comparison.txt output.log comparison.txt

echo "✅ All tests passed!"
