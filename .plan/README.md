# `.plan/` — resume contract

This directory is the **single source of truth** for the `simple-admin` build effort: what is decided, started, done, and blocked. It exists so any agent (or human) can resume work across sessions without re-reading the whole codebase.

> ⚠️ **STANDING RULE — every plan/phase in this repo uses this `.plan/` contract the same way.** Not optional, not just the first plan. Each new phase: add its steps to `progress.md`, append its decisions to `decisions.md`, create per-step `todo/*.md`, extend `PLAN.md`, and follow the per-step protocol below (mark `in_progress` → `done` + commit hash, verify each step).

> ⚠️ **STANDING DIRECTIVE — ALL AGENTS use caveman communication mode at all times** on this project: every response, every session, no reverting. Terse caveman — drop articles/filler, fragments OK, exact technical terms kept. Code, commits, and PRs are written normally. (Toggle phrase if ever needed: the user controls this.)

## Phases

- **Phase 1** (steps 0–13): MVP — done. See `todo/00..13-*.md`.
- **Phase 2** (steps p2-0..p2-13): cleanup/refactor + high-value features. See `todo/p2-*.md`.

## Files

- `README.md` — this file: purpose + the protocol below.
- `PLAN.md` — the full implementation plan (canonical spec). If reality diverges, update `PLAN.md`.
- `decisions.md` — append-only decision log (ADR-style): each locked decision + rationale + date.
- `progress.md` — status board: the 14 sequence steps (0–13) → status + one-line note + last-updated.
- `todo/NN-*.md` — one file per step: sub-task checklist, status, touched files, verification notes, follow-ups.

## Status vocabulary

`todo` · `in_progress` · `blocked` · `done`

## Protocol (every agent MUST follow)

1. **On resume:** read this `README.md` → `progress.md` (find first non-`done` step) → that step's `todo/NN-*.md` → `decisions.md`. Reconstructs context cheaply.
2. **Before starting a step:** set it `in_progress` in `progress.md` (with date); ensure its `todo/NN-*.md` exists with a sub-task checklist.
3. **While working:** check off sub-tasks, list touched files, note anything deferred.
4. **On finishing a step:** mark `done` in both files, record the commit hash, run the step's verification.
5. **On any new decision** (library choice, schema shape, scope change): append to `decisions.md`; if scope changes, update `PLAN.md`.
6. **`PLAN.md` is canonical** — keep it current rather than letting it rot.
