#!/bin/bash
# Compiles every function under netlify/functions/<name>/main.go into
# functions-build/<name>, the format Netlify expects for Go (Lambda-style)
# functions.
set -euo pipefail

OUT_DIR="functions-build"
rm -rf "$OUT_DIR"
mkdir -p "$OUT_DIR"

for dir in netlify/functions/*/; do
  name=$(basename "$dir")
  echo "Building function: $name"
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "$OUT_DIR/$name" "./$dir"
done

echo "Build complete: $(ls "$OUT_DIR")"
