#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if ! command -v gemini >/dev/null 2>&1; then
  echo "ERROR: gemini CLI not found in PATH"
  echo "Install or fix PATH before running this launcher."
  exit 1
fi

if [[ ! -f "bootstrap.txt" ]]; then
  echo "ERROR: bootstrap.txt not found at repo root: $ROOT"
  exit 1
fi

# Canonical DevOps entrypoint for MiddayCommander
# Usage:
#   ./bin/mdcgcli.sh "inspect repo and summarize architecture"
#   ./bin/mdcgcli.sh @docs/devops/some_prompt.txt
exec gemini --context @bootstrap.txt "$@"
