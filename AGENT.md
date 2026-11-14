# Agent Instructions — malarinv/forgejo

This repository is a **fork** of forgejo/forgejo kept in sync by forklift. Everything between the two markers below is generated from `.forklift/registry.yaml`: edit the registry and re-run `forklift render` rather than editing the block. Prose outside the markers is hand-maintained and preserved byte-for-byte.

<!-- GENERATED:FORKLIFT_START -->
<!-- Do not edit this block: `forklift render` regenerates it. The tables come from .forklift/registry.yaml; the situation playbook comes from templates/situations.yaml. -->
## Fork Identity & Remotes
- **Fork remote/branch**: `github/v17`
- **Upstream remote/branch**: `origin/v17.0/forgejo`
- **Expected fork commits**: `10`

## Fork-Owned Features
These are the paths this fork owns; prefer ours when they conflict with upstream, and register any new fork-owned path in `.forklift/registry.yaml` before relying on it.

| Feature | Description | Owned Paths |
| --- | --- | --- |
| admin-auth-sources-api | REST API for listing/reading/PATCHing OAuth2 auth sources (group_team_map, admin_group, group_claim_name, group_team_map_removal, quota maps). Consumed by Backstage repo-reconciler forgejoTeams.ts. | `modules/structs/auth_source.go`<br>`routers/api/v1/admin/auth_source.go`<br>`services/convert/auth_source.go`<br>`tests/integration/api_admin_auth_source_test.go` |
| package-cleanup-rules-api | REST CRUD for package cleanup rules at /api/v1/packages/{owner}/rules. Consumed by Backstage repo-reconciler ensurePackageCleanupRules. | `modules/structs/package_cleanup.go`<br>`routers/api/v1/packages/cleanup_rule.go`<br>`services/convert/package.go` |
| webhook-package-event-header | case HookEventPackage in HookEventType.Event() so X-Gitea-Event is 'package' for package webhooks; without it the Flux receiver cannot match package pushes. | `modules/webhook/type.go` |
| fork-build | Container build: Dockerfile.custom (v17 upstream rootless Dockerfile + RELEASE_VERSION default + FORGEJO_VERSION make arg) and GHA workflow building linux/arm64 to ghcr.io/malarinv/forgejo. | `Dockerfile.custom`<br>`.github/workflows/build.yaml`<br>`mise.toml` |

## Verification Commands
- **Typecheck**: `GOWORK=off go build ./modules/... ./routers/... ./services/... ./models/...`
- **Test**: `GOWORK=off go test -tags 'sqlite sqlite_unlock_notify' ./models/packages/... ./models/webhook/... ./modules/webhook/...`
- **Lint**: `GOWORK=off go vet $(go list ./modules/... ./routers/... ./services/... ./models/... | grep -v services/auth/source$)`

## Divergences & Intentional Removals
**Divergences (upstream-owned, take upstream):**

None registered.

**Intentional removals (deleted on purpose):**

None registered.

## Fork Maintenance Situations
The person asking for work on this fork describes **what they want**, not how the harness works. Match their request to a situation below, read that procedure, and follow it. Never ask them which command to run, which registry field to fill, or which files a feature owns — that is the procedure's work, not theirs.

| They say something like | What you do | Where the procedure is |
| --- | --- | --- |
| "set this fork up", "make this repo fork-aware", "we are forking this project" | draft the registry, install the harness, and hand back the worklist of what still needs claiming | the `forklift-init` skill |
| "I just added a feature", "make sure my new code survives the next sync" | register the feature so the next rebase cannot drop it | the `forklift-maintain` skill, § *Register a feature* |
| "sync with upstream", "pull in upstream's changes", "rebase onto upstream" | size the conflict surface, rebase, resolve, gate, and push | the `sync-upstream` skill |
| "is the fork healthy", "what drifted", "does the registration still match the code" | audit the registration against the tree and report the findings | the `forklift-doctor` skill |
| "the rebase went wrong", "features got dropped", "the history is too tangled to rebase" | triage the history and rebuild it topically | the `forklift-maintain` skill, § *Rebuilding the history* |
<!-- GENERATED:FORKLIFT_END -->

## Repository Specific Notes

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
