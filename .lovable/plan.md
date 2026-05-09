## Goal

Break the 2,216-line `.github/workflows/ci.yml` into small, focused, reusable pieces so no single file exceeds ~150 lines and each piece has one responsibility. We do this incrementally over **5 steps**, one per turn. After each step you review and say "next" before I move on.

## Approach (technical)

GitHub Actions does not let one workflow span multiple files, but it gives us two tools:

1. **Reusable workflows** (`on: workflow_call`) — extract a group of related jobs into `.github/workflows/_<name>.yml`, then call them from `ci.yml` with `uses: ./.github/workflows/_<name>.yml`.
2. **Composite actions** (`.github/actions/<name>/action.yml`) — extract repeated step sequences (checkout + setup-go + cache, misspell install, lint baseline diff, etc.) into a single referenceable action.

End state: `ci.yml` becomes a thin orchestrator (~150 lines) that wires together reusable workflows + composite actions. Each extracted file stays under ~150 lines.

## Target layout

```text
.github/
  actions/
    setup-go-cached/action.yml         # checkout + setup-go + GOMODCACHE/GOCACHE cache
    install-misspell/action.yml        # pinned v0.3.4 install
    install-golangci-lint/action.yml   # pinned v1.64.8 install
    baseline-diff/action.yml           # per-linter baseline-vs-head diff
  workflows/
    ci.yml                             # orchestrator only (~150 lines)
    _diagnostics.yml                   # diag, sha-check, ci-errors-digest
    _lint.yml                          # spell-check, lint, lint-script-tests
    _lint-baseline.yml                 # lint-baseline-guard, lint-baseline-diff
    _repo-policy.yml                   # cmd-naming, legacy-refs, deploy-layout,
                                       #   constants-naming, constants-collision,
                                       #   golden-allow-leak, generate-check
    _security-tests.yml                # vulncheck, json-snapshot-fast, test,
                                       #   full-suite-guard
    _smoke.yml                         # installer-smoke, installer-smoke-windows
    _build.yml                         # build, build-summary
    _summaries.yml                     # pr-summary, test-summary
```

Job dependencies (`needs:`) and SHA-dedup passthrough are preserved by passing inputs/outputs between reusable workflows.

## The 5 steps

**Step 1 — Composite actions (foundation).**
Create `.github/actions/setup-go-cached`, `install-misspell`, `install-golangci-lint`, `baseline-diff`. Refactor existing jobs in `ci.yml` in place to use them. No job graph changes yet. This shrinks `ci.yml` by ~300 lines and proves the pattern.

**Step 2 — Extract `_diagnostics.yml` + `_summaries.yml`.**
Move `diag`, `sha-check`, `ci-errors-digest`, `pr-summary`, `test-summary` into two reusable workflows. Wire them from `ci.yml` with `uses:` and pass through the SHA-dedup output. Lowest-risk extraction (no `needs:` from other jobs depend on their internals).

**Step 3 — Extract `_lint.yml` + `_lint-baseline.yml`.**
Move `spell-check`, `lint`, `lint-script-tests`, `lint-baseline-guard`, `lint-baseline-diff`. Largest single win (~1,000 lines out of `ci.yml`). Verify baseline-diff job still reads the same `main` baseline correctly across the workflow_call boundary.

**Step 4 — Extract `_repo-policy.yml` + `_smoke.yml`.**
Move the 7 repo-policy checks and both installer-smoke jobs. These already shell out to `.github/scripts/*.sh` so the YAML is thin — easy lift.

**Step 5 — Extract `_security-tests.yml` + `_build.yml`; finalize orchestrator.**
Move `vulncheck`, `json-snapshot-fast`, `test`, `full-suite-guard`, `build`, `build-summary`. Then trim `ci.yml` to just `on:` triggers + concurrency + the `uses:` graph. Final pass: confirm every file ≤ ~150 lines, all `needs:` edges preserved, full pipeline green on a test PR.

## Guardrails (carried through every step)

- Pinned tool versions stay pinned (`golangci-lint@v1.64.8`, `govulncheck@v1.1.4`, `misspell@v0.3.4`).
- SHA-dedup passthrough gate is preserved end-to-end.
- `cancel-in-progress: false` for `release/**` stays intact (per `spec/02-app-issues/18`).
- No `cd` in CI — `working-directory` only.
- Each PR/step is independently revertable.
- After each step I list what's done and what remains, and wait for your "next".

## Verification per step

1. `actionlint` (or `yamllint`) on every changed workflow file.
2. Push to a throwaway branch, confirm the full job matrix renders identically in the GitHub UI (same job names, same `needs:` edges).
3. Re-check line counts — fail the step if any file > 150 lines.

Ready to execute Step 1 on your go.