---
name: sync-upstream
description: Rebase malarinv/forgejo (forgejo-fork) fork/v17 against forgejo/forgejo (forgejo) upstream/v17.0/forgejo. Handles conflict resolution automatically for trivial cases and escalates semantic conflicts. Read this skill before running forklift sync.
---
<!-- forklift v1 -->

# Sync Upstream (forgejo-fork ← forgejo)

Rebase `fork/v17` against `upstream/v17.0/forgejo` (forgejo). Run `forklift sync` to execute. This skill provides the knowledge you need to resolve conflicts correctly during the rebase.

## Script

```bash
bun /var/home/user/Work/forklift/bin/forklift sync
```

Git operations run through the command above. You resolve conflicts using the decision tree below, then signal the script to continue or abort.

## Conflict Walk

Read a conflict from git's three-way index, not from the conflict-marker file: markers hide which side changed what and cannot show what the two sides share.

**Before the rebase starts**, reconcile the branch picture against what the registry claims — this is the macro intent every per-commit decision below serves:

```bash
git diff --stat origin/v17.0/forgejo...v17
```

**Extract the three stages** of each conflicted path into a scratch directory (keep the extension — parsers key on it):

```bash
p=<conflicted-path>; ext=${p##*.}; d=$(mktemp -d)
git ls-files -u -- "$p"                 # which stages actually exist
git show :1:"$p" > "$d/base.$ext"       # common ancestor
git show :2:"$p" > "$d/ours.$ext"       # so-far-rebased series, starting from forgejo
git show :3:"$p" > "$d/theirs.$ext"     # the fork commit being replayed
```

During a rebase the roles swap: `:2:` (ours) is the rebased series, `:3:` (theirs) is your feature commit.

**Missing stages.** For add/add, delete/modify and rename conflicts a stage is absent — `git ls-files -u` lists only the stages git holds. Read an absent stage as an empty file and take the decision tree's semantic branch (escalate); never reconstruct the missing side from marker text.

**Read each side against base:**

```bash
git diff --no-index --color=never "$d/base.$ext" "$d/ours.$ext"    # upstream's change
git diff --no-index --color=never "$d/base.$ext" "$d/theirs.$ext"  # the feature's change
```

`--no-index` exits 1 when the files differ — the normal case, not an error.

This fork registers `difftastic` as its structural reader (`harness.checks.semantic_drift.tool: difftastic`). Prefer it: it reports syntax-level change and hides upstream reformatting.

```bash
difft --color never "$d/base.$ext" "$d/ours.$ext"
difft --color never "$d/base.$ext" "$d/theirs.$ext"
```

**Prerequisites.** `difft` is not on PATH by default — confirm it through the fork's mise setup before the first read, and confirm the language tooling the fork's own verification needs while you are there (`GOWORK=off go build ./modules/... ./routers/... ./services/... ./models/...` fails on a missing language server just as surely as the diff fails on a missing binary):

```bash
mise exec -- difft --version
```

When it is absent, add it to the fork's `mise.toml` `[tools]` as `"github:Wilfred/difftastic" = "latest"` (GitHub releases index; pin an exact tag such as `= "0.71.0"` when reproducibility matters) and run `mise install`; the same `[tools]` block is where a fork pins the language servers, formatters and test runners its `GOWORK=off go build ./modules/... ./routers/... ./services/... ./models/...` / `GOWORK=off go test -tags 'sqlite sqlite_unlock_notify' ./models/packages/... ./models/webhook/... ./modules/webhook/...` need. Re-render the skill after adding the key.

`difft` carries its own tree-sitter grammars, so no per-language parser install is required for the languages it supports; a file it cannot parse degrades to a plain line diff, so treat output with no structure as line-based.

`difft` is a read-only analyst — its output is analysis, never a patch; do not write it to the file, and never treat it as proof the result compiles.

**Noise filter.** When `base → ours` differs from base only in whitespace or reformatting, the hunk is a formatter artifact, not a semantic clash: keep the feature's content from `:3:`, skip ESCALATE, and prove it by compiling. When `base → ours` carries a real change to a file listed in § Owned Symbols, it is a MERGE PATH conflict: take upstream's structure and graft the named symbols into it.

**Gate each paused commit.** After resolving one commit's conflicts and before `git rebase --continue`, run the protecting tests of every feature whose owned paths you just touched:

```bash
GOWORK=off go test -tags 'sqlite sqlite_unlock_notify' ./models/packages/... ./models/webhook/... ./modules/webhook/... <protecting_tests for the touched features>
```

A red test here names the resolution that broke the feature; waiting for the end-of-rebase suite attributes the break to dozens of commits.

## Forgejo fork conflict semantics

- The fork is a feature fork: 3 behavioral features graft small hunks into 2 shared files
  (routers/api/v1/api.go, modules/webhook/type.go); everything else is fork-owned files.
  Upstream rewrites line numbers every release, so conflicts in the shared files are context
  drift, not semantic.
- routers/api/v1/api.go graft points:
  - /rules group: inside m.Group("/packages/{username}", ...), after m.Post("/-/unlink", ...)
    and before m.Get("/", packages.ListPackages).
  - /auth-sources group: inside m.Group("/admin", ...), after the /unadopted group, before the
    admin group's closing `}, tokenRequiresScopes(auth_model.AccessTokenScopeCategoryAdmin), ...)` line.
- modules/webhook/type.go graft point: inside func (h HookEventType) Event(), immediately after
  `case HookEventRelease: return "release"`.
- v17 renamed nothing the features call (verified against v17.0.1 source): auth.GetSourceByID,
  auth.UpdateSource, db.Find[auth.Source], oauth2.Source fields (GroupTeamMap, AdminGroup,
  GroupClaimName, GroupTeamMapRemoval, quota fields) and the packages cleanup-rule model funcs
  all exist with the same signatures. A conflict inside routers/api/v1/admin/auth_source.go or
  cleanup_rule.go is upstream context change only — take upstream structure, re-apply handlers.
- services/convert/package.go: v17 added two PackageCleanupRule fields (KeepLastDownloadDays,
  RemoveLastDownloadDays). ToPackageCleanupRule must map them when re-applying the fork's
  converter addition, or the API returns zeroed values (silent data bug, not a compile error).

## Tag Strategy (Remote Backup)

Use `--tag` to create lightweight annotated tags as remote backups of known-working states. These tags are pushed to github when combined with `--push` and `--follow-tags`.

### Tag format

| Tag | Commit | Created | Purpose |
|-----|--------|---------|---------|
| `sync/base/YYYY-MM-DD` | pre-sync HEAD | Before rebase starts | Rollback point if sync fails or introduces regressions |
| `sync/YYYY-MM-DD` | post-sync HEAD | After rebase + typecheck pass | New known-working state after successful sync |
| `sync/YYYY-MM-DD-N` | — | Collision handling | If multiple syncs land on the same day, `-N` suffix increments |

### Usage

```bash
# Full sync with tags and push
bun /var/home/user/Work/forklift/bin/forklift sync --push --tag

# Tags created locally; push separately
bun /var/home/user/Work/forklift/bin/forklift sync --tag
# ... later ...
git push github v17 --force-with-lease --follow-tags
```

### Why tags instead of just sync-state.json?

- `sync-state.json` lives at `~/.omp/forklift/forgejo-sync-state.json` — lost if the machine is replaced
- Tags on github survive independently of local state
- Tags survive `git gc` and reflog expiry better than loose refs
- Easy historical diff: `git log --oneline sync/2026-06-01..v17`

### Rollback with tags

If a pushed sync introduces regressions:

```bash
# Find the last good pre-sync tag
git tag -l 'sync/base/*' | sort | tail -1
# Reset to it
git reset --hard sync/base/2026-06-01
# Force push the rollback
git push github v17 --force-with-lease
```

## Post-Rebase Checklist

After all conflicts are resolved and compilation passes, perform these bookkeeping steps before force-pushing:

- **Converge the commit count and the generated docs in one step.** Run
  ```bash
  bun /var/home/user/Work/forklift/bin/forklift sync --rebaseline
  ```
  It records the fork's commit count in `.forklift/registry.yaml`, re-renders every generated
  artifact the registry describes (SKILL.md, `AGENT.md`, the fork-commits table in
  `.forklift/README.md`), and lands them as one atomic commit — the count it records is the value the
  range will have once that commit lands, so `forklift doctor` check 5 passes afterwards.
  Do **not** hand-edit `expected_fork_commit_count` and do **not** re-render into a second
  commit: the count is self-referencing, and a second render leaves the fork doc one behind.
  If you already committed something by hand, run `bun /var/home/user/Work/forklift/bin/forklift sync --rebaseline-only` — it
  converges the bookkeeping without fetching or rebasing.
- **Update feature registry.** If new files were added, removed, or moved, update `.forklift/registry.yaml` and re-run the rebaseline above (it re-renders).
- **Update divergence list.** If a previously divergent file is now owned by forgejo-fork, update both lists.

## Feature Stack

These are forgejo-fork-owned files. Conflicts here prefer ours unless upstream's change is semantic.

| Feature | Owned files |
|---------|------------|
| admin-auth-sources-api | `modules/structs/auth_source.go`, `routers/api/v1/admin/auth_source.go`, `services/convert/auth_source.go`, `tests/integration/api_admin_auth_source_test.go` |
| package-cleanup-rules-api | `modules/structs/package_cleanup.go`, `routers/api/v1/packages/cleanup_rule.go`, `services/convert/package.go` |
| webhook-package-event-header | `modules/webhook/type.go` |
| fork-build | `Dockerfile.custom`, `.github/workflows/build.yaml`, `mise.toml` |

> **Note:** The § Owned Symbols table below is generally more current than this § Feature Stack, because new shared-file symbols are registered immediately while high-level feature rows may lag. After any feature merge, audit both tables.

## Owned Symbols in Shared Files

When a conflict occurs in a file shared with upstream, the Feature Stack tells you "prefer ours"
but doesn't tell you *which lines are ours*. This section lists the exact symbols each feature
owns inside shared files.

### api.go (routers/api/v1/api.go)

| Feature | Symbol | Location | Description |
|---------|--------|----------|-------------|
| m.Group("/auth-sources", ...) | routers/api/v1/api.go, admin group tail (after /unadopted, before the admin group's closing tokenRequiresScopes line) | Route registration for /api/v1/admin/auth-sources (list/get/PATCH) |

### api.go (routers/api/v1/api.go)

| Feature | Symbol | Location | Description |
|---------|--------|----------|-------------|
| m.Group("/rules", ...) | routers/api/v1/api.go, packages group (after m.Post("/-/unlink", ...), before m.Get("/", packages.ListPackages)) | Route registration for /api/v1/packages/{owner}/rules CRUD |

### type.go (modules/webhook/type.go)

| Feature | Symbol | Location | Description |
|---------|--------|----------|-------------|
| case HookEventPackage | modules/webhook/type.go, Event() switch (immediately after case HookEventRelease) | Two-line case mapping HookEventPackage to the 'package' event header |

## Upstream Divergences (take upstream)

These forgejo files exist but forgejo-fork intentionally doesn't override them. Always take upstream:

> **WARNING:** The divergence list exists because these files are considered upstream-owned. If you add an forgejo-fork feature that modifies one of these files, the feature will be silently lost on the next sync (the decision tree will "prefer upstream"). You must either:
> 1. Move the feature out of these files, or
> 2. Add the file to the Feature Stack registry and change the divergence list entry to note the dual ownership.

## Obsoleted Fixes

No upstream-obsoleted fixes are currently tracked.

## Intentional Removals (deleted vs upstream)

These upstream paths are intentionally absent from forgejo-fork. `forklift sync --verify` fails if any other upstream file goes missing.

None registered.

## Escalation Format

When escalating, output exactly:

```
ESCALATE
file: <path>
reason: <one sentence: what upstream changed and which forgejo-fork feature it affects>
upstream: <brief description of upstream's version>
ours: <brief description of forgejo-fork's version>
question: adapt our feature to upstream's change, or drop the upstream change?
```

## Sync Options

`forklift sync` supports additional flags beyond the basic rebase:

| Flag | Behavior |
|------|----------|
| `--push` | Push github/v17 after successful rebase |
| `--dry-run` | Preview-only: shows what would happen without modifying `v17` |
| `--continue` | Resume a previously started rebase (run checks, push) |
| `--verify` | Run post-rebase verification suite (conflict markers, symbols, typecheck) |
| `--rebaseline` | After the rebase: record the fork's commit count, re-render the generated artifacts and land them as one atomic commit (the bookkeeping an agent used to do by hand) |
| `--rebaseline-only` | Converge the commit count and the generated artifacts on their own — never fetches, never rebases |
| `--tag` | Create lightweight tags for pre-sync and post-sync commits as remote backups (pushed with `--follow-tags` when `--push` is also set) |

## After Resolution

Once all conflicts are resolved:

- **Verify no conflict markers remain:**
  ```bash
  git grep -n '<<<<<<<'
  ```

- **Run the type check (not just lint):**
  ```bash
  GOWORK=off go build ./modules/... ./routers/... ./services/... ./models/...
  ```
  This catches structural breakage (truncated objects, orphaned references) that lint misses. Ignore pre-existing errors in dependency directories or upstream-owned files.

- **Run the full verification suite — this is the check that matters:**
  ```bash
  bun /var/home/user/Work/forklift/bin/forklift sync --verify
  ```
  It gates on exit code and covers, in order: conflict markers, symbol health,
  commit survival, protecting tests (feature behavioral contracts), **feature
  coverage** (every behavioral feature names a `protecting_tests` entry),
  **import liveness** (every owned source has at least one importer), footprint,
  type check, and the deletion audit.

  Import liveness is the check that catches a feature whose files are present and
  registered but whose wiring was dropped: nothing imports it, so its own file
  still greps fine while the capability is dead.

- **Verify forgejo-fork-owned symbols survive:** prefer `forklift sync --verify`'s
  symbol-health step over ad-hoc greps; the § Owned Symbols table is partly
  auto-derived and its entries are local variable names, so a hand grep against it
  proves nothing about a feature's reachability.

- **Verify no commits were silently dropped:**
  ```bash
  git rev-list --count origin/v17.0/forgejo..v17
  ```
  Compare against `expected_fork_commit_count` in the registry. Treat a moved
  number as a *symptom only*: it names no feature. When it moves, use the
  per-feature coverage and liveness results to find which capability actually
  stopped running.

- Verify the binary version string: Dockerfile.custom bakes RELEASE_VERSION into the version
  output via FORGEJO_VERSION; `gitea --version` must print `17.0.1-fork`, not a git-describe
  hash, or the GHA shallow-clone override regressed.

- **Update .forklift/README.md § Last Sync Point.**

- **Push:**
  ```bash
  git push github v17 --force-with-lease
  ```
  If the push fails with `remote: fatal: did not receive expected object <sha>`,
  see § Push Failure: Fork-Network Missing Objects below.

## Pre-Rebase Registry Validation

`forklift sync` validates every file touched by fork commits against the registry *before* starting the rebase. If any file is unregistered (not in features or divergences), the script aborts:

```
ERROR: Unregistered file in fork commit <sha>: <path>
  This file is not in the registry's § features or § divergences.
  Add it to the registry, then regenerate SKILL.md and re-run.
```

This prevents silent loss of features on the next sync.

## Post-Rebase Automated Verification

Run `forklift sync --verify` after a completed rebase to run the full verification suite:

| Check | Method | Fail behavior |
|-------|--------|---------------|
| Conflict markers | `git grep -n '<<<<<<<' -- 'modules/**/*.go' 'routers/**/*.go' 'services/**/*.go' 'models/**/*.go' 'Dockerfile*' '.github/workflows/*.yaml'` | Hard fail |
| Symbol health | `grep` each forgejo-fork-owned symbol from § Owned Symbols | Report PASS/MISSING |
| Commit survival | `git rev-list --count origin/v17.0/forgejo..v17` | Report expected vs actual |
| Protecting tests | `GOWORK=off go test -tags 'sqlite sqlite_unlock_notify' ./models/packages/... ./models/webhook/... ./modules/webhook/... <protecting_tests>` per registered file | Hard fail |
| Feature coverage | every behavioral feature declares `protecting_tests` | Hard fail |
| Import liveness | owned sources resolved against parsed import specifiers | Hard fail |
| Type check | `GOWORK=off go build ./modules/... ./routers/... ./services/... ./models/...` | Report pass/fail |

Output is a structured JSON block plus human-readable summary.

## Push Failure: Fork-Network Missing Objects

**Symptom:**
```
remote: fatal: did not receive expected object 2a0da9f7653a83fa51b38b36436aea2eb3dc7fe8
error: remote unpack failed: index-pack failed
```

**Root cause:** GitHub fork networks share an object database between the parent repo
and the fork. During push negotiation,
the fork advertises objects it can see through the alternate as `have`, so the client
builds a thin pack with deltas against those bases. But when `receive-pack` runs
server-side it looks up the bases in the *fork's own* object DB (which doesn't see
through the alternate during writes) and finds them missing → unpack failure.

Typically triggered after the parent repo is repacked, force-pushed, or GC'd in a way
that orphans objects from the fork's view but leaves stale `have` advertisements.

**Fix:** push each missing object as a temporary ref on github, forcing the fork to
copy it into its own object database. Loop until the target push succeeds, then delete
the temp refs.

```bash
for i in $(seq 1 15); do
  out=$(git push github v17 --force-with-lease 2>&1)
  missing=$(echo "$out" | grep -oE 'expected object [0-9a-f]{40}' | awk '{print $3}' | head -1)
  [ -z "$missing" ] && { echo "$out" | tail -3; break; }
  echo "Pinning missing object: $missing"
  git fetch origin "$missing" 2>&1 | tail -1
  git push github "$missing:refs/heads/_owns_${missing:0:8}" 2>&1 | tail -1
done

# Cleanup
refs=$(git ls-remote github '_owns_*' | awk '{print $2}' | sed 's|refs/heads/||')
[ -n "$refs" ] && git push github --delete $refs
```

Each iteration unblocks one delta base. Three to five iterations is typical for a
multi-week sync gap; sometimes only one. The temporary `_owns_<sha8>` branches are
harmless if forgotten but should be deleted when the push completes.

## Rollback

If the rebase introduces regressions, `forklift sync` records the pre-sync HEAD in `~/.omp/forklift/forgejo-sync-state.json` before touching git.

**Automatic revert (recommended):**
```bash
bun /var/home/user/Work/forklift/bin/forklift sync --revert
```
This resets `fork/v17` to the exact pre-sync commit without touching the reflog.
It only works if the last sync was not yet marked `completed`; completed syncs must be reverted manually to avoid accidentally discarding upstream changes.

**Check sync history:**
```bash
bun /var/home/user/Work/forklift/bin/forklift sync --status
```
Shows the last sync attempt, pre-sync HEAD, upstream base, and last 5 log entries from `~/.omp/forklift/forgejo-sync-log.jsonl`.

**Manual fallback:**
```bash
git rebase --abort          # during rebase
git reset --hard ORIG_HEAD  # after a bad rebase completed (if you didn't push)
git push github v17 --force-with-lease
```

The append-only log at `~/.omp/forklift/forgejo-sync-log.jsonl` records every attempt, so you can correlate regression patterns with specific upstream bases.
