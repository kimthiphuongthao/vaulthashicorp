#!/usr/bin/env bash
set -euo pipefail

# Create and switch to a new feature branch with a conventional prefix.
# Usage:
#   ./scripts/git-new-feature-branch.sh short-slug

slug="${1:-}"
if [[ -z "$slug" ]]; then
  echo "ERROR: slug required"
  echo "Usage: $0 short-slug"
  exit 2
fi

branch="feature/${slug}"

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "ERROR: not a git repo"
  exit 2
fi

git checkout -b "$branch" 2>/dev/null || git checkout "$branch"

echo "On branch: $branch"
