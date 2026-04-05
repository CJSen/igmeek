# igmeek Init And File Matching Design

**Goal:** Improve `igmeek init` so first-time setup collects both GitHub token and an initial repository, runs an immediate full `sync`, and reports both sync results and storage paths. Also make file-based issue lookup work reliably for absolute and relative paths across Windows, Linux, and macOS.

**Context**

Current behavior has two gaps:

- `igmeek init` only saves the token and does not collect a repository or run an initial sync.
- File lookup for commands that accept `file` currently relies on exact `FilePath` string equality, so absolute paths and relative paths fail to match the same indexed issue.

These gaps are visible in the current implementation:

- `cmd/init.go` only prompts for token and saves config.
- `cmd/sync.go` requires `CurrentRepo` to already exist.
- `internal/index/index.go` matches files only with exact `FilePath == path` logic.

## Decision Summary

Use the minimal-correct approach:

- Extend `init` instead of redesigning config storage.
- Accept both `owner/repo` and GitHub repository URLs as the initial repo input.
- Save token and repo first, then run one automatic full sync.
- If sync fails, keep the saved config and return an actionable error.
- Keep path matching conservative: exact path match first, basename fallback second, ambiguity fails with guidance.
- Add similarity suggestions only as hints, never as automatic matching.

## Design 1: Init Flow

`igmeek init` will become the first-run setup entrypoint for both authentication and repository bootstrap.

### User Interaction

The command prompts in this order:

1. GitHub token
2. Initial repository

The repository input accepts:

- `owner/repo`
- `https://github.com/owner/repo`
- `http://github.com/owner/repo`

The input is normalized to canonical `owner/repo` form before saving.

### Config Write Behavior

After validation:

- save `token`
- add the repo to `repos` if it does not already exist
- set `current_repo` to that repo

This preserves the existing config model and avoids broad storage changes.

### Automatic Sync

After saving config, `init` runs one full `sync` using the selected repo.

On success, the command prints:

- current repository
- synced issue count
- synced label count
- global config path
- current repo data directory path

### Failure Semantics

If token and repo are valid but the automatic sync fails:

- keep the saved config
- return an error
- make the message explicit that token and repo were saved successfully
- tell the user they can retry with `igmeek sync`

This matches the chosen behavior of preserving setup progress instead of rolling everything back.

## Design 2: File Matching Rules

Commands that resolve an issue from a `file` argument should no longer depend on the raw input path string exactly matching the stored index path.

### Cross-Platform Path Normalization

Before matching, normalize the user input path with platform-aware path handling.

Requirements:

- support relative and absolute paths
- support Windows separators (`\\`) and Unix separators (`/`)
- compare basenames using normalized paths
- avoid assumptions tied to a single operating system

The matching logic must work consistently on Windows, Linux, and macOS.

### Matching Priority

Use this resolution order:

1. Exact normalized path match
2. Basename match using the last file name segment
3. If basename match returns exactly one result, use it
4. If basename match returns multiple results, fail with an ambiguity error

This preserves precision when possible while fixing absolute-path vs relative-path mismatches.

### Error Behavior

If no match is found, return:

- `未找到对应文件名的 issue 映射，先执行 sync 或显式传入 issue_number`

If multiple same-name files match, return:

- `存在多个同名文件，请使用 igmeek update <num> <file>`

### Suggestions, Not Auto-Match

When no match is found, compute a small set of similar candidate names from the local index and show them as suggestions.

When multiple same-name matches are found, print the conflicting candidates as suggestions too.

Examples:

- `相近文件：foo.md, foo-test.md`
- `候选：posts/foo.md -> #12, drafts/foo.md -> #37`

Important boundary:

- similarity only helps the user decide
- it must never silently choose a target issue

## Components Affected

### `cmd/init.go`

Needs to:

- prompt for repo in addition to token
- normalize repo input
- update config with repo membership and current repo
- call sync after saving config
- print sync result and storage paths
- return preserved-config guidance on sync failure

### `cmd/sync.go`

Likely needs a reusable helper path so `init` can invoke the same sync behavior without duplicating repository parsing and result formatting logic.

### `internal/config/config.go`

May need small helper additions for:

- repo normalization support location, if kept near config/bootstrap code
- clearer path reporting utilities

No config schema redesign is required.

### `internal/index/index.go`

Needs the main lookup behavior change:

- normalized exact-path comparison
- basename fallback lookup
- ambiguity detection
- suggestion support

This is the primary fix for the current `file` matching bug.

## Data Flow

### Init

1. Read token input
2. Read repo input
3. Normalize repo to `owner/repo`
4. Load or initialize config
5. Save token and repo config
6. Resolve repo data directory
7. Run full sync
8. Print result and storage paths

### File Lookup

1. Receive `file` argument
2. Normalize path according to current platform rules
3. Attempt exact normalized path lookup
4. If not found, extract basename and search by basename
5. If one match, use it
6. If none, return not-found error plus similar-name suggestions
7. If multiple, return ambiguity error plus candidate list

## Error Handling

### Invalid Repo Input

Reject repo input that cannot be normalized into `owner/repo` and return a clear validation error before saving config.

### Sync Failure During Init

Return an error that includes both facts:

- config was saved successfully
- sync failed and should be retried manually

### Ambiguous File Lookup

Do not guess. Require the user to switch to explicit `issue_number` mode.

## Testing Strategy

Add tests for both behavior changes.

### Init Tests

Cover:

- token + `owner/repo` input saves config and triggers sync
- token + GitHub URL input is normalized to `owner/repo`
- sync failure keeps saved config and returns the correct guidance
- success output contains counts and storage paths

### Index Matching Tests

Cover:

- exact path match still works
- absolute input matches stored relative path by basename fallback
- relative input matches stored absolute path by basename fallback
- Windows-style path input is handled correctly
- basename duplicate returns ambiguity error
- not-found case returns similarity suggestions

## Out Of Scope

This design does not include:

- redesigning `issues_num_name.json`
- automatic conflict resolution
- fuzzy auto-binding of files to issues
- changing the overall repo management model

## Rationale

This approach solves the concrete problems raised by the user with the smallest reliable change set:

- `init` becomes useful on first run instead of requiring a second manual repo setup step.
- the first-run experience ends with real synced data and visible storage locations.
- file lookup becomes practical for real CLI use where users naturally mix absolute and relative paths.
- cross-platform path handling is treated as a first-class requirement instead of an accidental side effect.
