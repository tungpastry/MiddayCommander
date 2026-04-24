#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "[INFO] repo=$ROOT"

echo "[CHECK] git status"
git status --short

echo "[CHECK] go version"
go version

echo "[CHECK] go test ./..."
go test ./...

echo "[CHECK] go vet ./..."
go vet ./...

echo "[CHECK] make build"
make build

echo "[OK] DevOps validation completed"
