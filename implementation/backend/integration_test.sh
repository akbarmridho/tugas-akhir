#!/bin/bash

set -euo pipefail

COVERAGE_DIR=.cover
COVERAGE_FILE=cover.out

# Clean old coverage files
rm -rf $COVERAGE_DIR
mkdir -p $COVERAGE_DIR

# Run tests package-by-package
for pkg in $(go list ./...); do
    go test -coverprofile="$COVERAGE_DIR/$(echo "$pkg" | tr / -).cover" -covermode=atomic "$pkg"
done

# Combine all .cover files into one
echo "mode: atomic" > "$COVERAGE_FILE"
tail -q -n +2 $COVERAGE_DIR/*.cover >> "$COVERAGE_FILE"

# Show total coverage in terminal
go tool cover -func="$COVERAGE_FILE"

# Generate HTML coverage report
go tool cover -html="$COVERAGE_FILE" -o cover.html

echo "Done. Combined coverage available in $COVERAGE_FILE and cover.html."
