#!/usr/bin/env bash
set -euo pipefail

# End-of-day helper: create a feature branch if needed, commit all changes, and push to origin.
# Usage:
#   ./scripts/eod-push.sh "your commit message"
#   BRANCH=feature/my-branch ./scripts/eod-push.sh "msg"

msg="${1:-}"
if [[ -z "$msg" ]]; then
  echo "ERROR: commit message required"
  echo "Usage: $0 \"message\""
  exit 2
fi

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "ERROR: not a git repo"
  exit 2
fi

# Ensure origin exists
if ! git remote get-url origin >/dev/null 2>&1; then
  echo "ERROR: remote 'origin' is not configured"
  exit 2
fi

current_branch="$(git rev-parse --abbrev-ref HEAD)"

desired_branch="${BRANCH:-}"
if [[ -z "$desired_branch" ]]; then
  if [[ "$current_branch" == "main" || "$current_branch" == "master" ]]; then
    desired_branch="feature/eod-$(date +%Y%m%d-%H%M%S)"
  else
    desired_branch="$current_branch"
  fi
fi

if [[ "$current_branch" != "$desired_branch" ]]; then
  if git show-ref --verify --quiet "refs/heads/$desired_branch"; then
    git checkout "$desired_branch"
  else
    git checkout -b "$desired_branch"
  fi
fi

# No changes?
if git diff --quiet && git diff --cached --quiet; then
  if [[ -z "$(git status --porcelain)" ]]; then
    echo "No changes to commit. Pushing branch anyway..."
    git push -u origin HEAD
    exit 0
  fi
fi

# Stage everything (including new files)
git add -A

# Commit only if there is something staged
if git diff --cached --quiet; then
  echo "Nothing staged to commit."
else
  git commit -m "$msg"
fi

# Push and set upstream if needed
git push -u origin HEAD

echo "Done: pushed $(git rev-parse --abbrev-ref HEAD)"
