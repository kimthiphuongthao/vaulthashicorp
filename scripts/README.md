# Git helper scripts

## End-of-day push

```bash
chmod +x scripts/*.sh
./scripts/eod-push.sh "chore: end of day sync"
```

- If you're on `main`/`master`, it auto-creates a new branch like `feature/eod-YYYYMMDD-HHMMSS`.
- Otherwise it commits + pushes the current branch.

## Create feature branch

```bash
./scripts/git-new-feature-branch.sh vault-plugin-hardening
```
