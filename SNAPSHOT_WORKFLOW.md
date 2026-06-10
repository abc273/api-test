# Snapshot Workflow

This project includes a Git snapshot workflow so you can always see what changed and roll back safely.

## Commands

Create a snapshot before editing:

```bash
./scripts/git-snapshot create "before fixing login redirect"
```

List all snapshots:

```bash
./scripts/git-snapshot list
```

Show one snapshot in detail:

```bash
./scripts/git-snapshot show latest
./scripts/git-snapshot show snapshot/20260610-131500-before-fixing-login-redirect
```

Compare your current workspace against a snapshot:

```bash
./scripts/git-snapshot diff latest
```

Restore back to a snapshot:

```bash
./scripts/git-snapshot restore latest
```

## Recommended Workflow

1. Before asking Codex to change code, create a named snapshot.
2. Let Codex make changes.
3. Check the changes with `git diff` or `./scripts/git-snapshot diff latest`.
4. If the result is good, commit normally.
5. If the result is bad, run `./scripts/git-snapshot restore latest`.

## Notes

- Every snapshot is both a real Git commit and a tag under `snapshot/...`.
- Snapshot restore automatically creates a safety snapshot of your current state first.
- `create` stages tracked and untracked files. Ignored files such as `.env` are still ignored.
