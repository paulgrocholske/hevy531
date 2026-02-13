# AGENTS.md

Guidance for coding agents working in this repository.

## Project Purpose

This is a Go CLI that generates a 4-week 5/3/1 BBB lifting program, then either:
- exports it to CSV, or
- syncs routines to the Hevy API.

The entrypoint is interactive (`go run .`), not flag-driven.

## Tech Stack

- Language: Go (`go.mod` currently declares `go 1.25.6`)
- No external Go dependencies (std lib only)
- Main binary name in repo context: `531bbb`

## Repository Map

- `main.go`: top-level flow (prompt -> program generation -> CSV export or Hevy sync)
- `config/config.go`: domain constants/types, default config, training max + rounding helpers
- `prompt/prompt.go`: all CLI input/output prompts and validation
- `program/program.go`: deterministic program generation (weeks, sets, BBB, accessories)
- `export/csv.go`: CSV writer
- `hevy/client.go`: Hevy API client + exercise mapping/aliases + lbs->kg conversion
- `hevy/converter.go`: converts generated program days into Hevy routine payloads

## Core Behavior Invariants

Preserve these unless a task explicitly changes them:

1. Program shape is fixed at 4 weeks x 4 training days (`16` days total).
2. Weeks 1-3 include warmups and working sets; week 4 uses deload scheme only.
3. Weight math is based on training maxes and rounded to nearest 5 lb.
4. BBB work is a single entry of `5x10` at configured BBB percentage.
5. Accessories are `5x10` with no weight/percentage (blank in CSV; no weight in Hevy sets).
6. Routine titles are formatted as `531 BBB W{week}D{day} - {MainLift}`.
7. Hevy sync uses week folders named `531 BBB Week {n}` and updates existing routines by title.
8. Hevy routine updates must omit `folder_id` (allowed on create, not on update).
9. Hevy sync includes retry/backoff behavior for rate-limit style failures.

## Development Workflow

Run from repo root:

1. `go run .`
2. `go build ./...`
3. `go test ./...`
4. `gofmt -w .`

Notes:
- Current repo has no committed tests yet; add tests for behavior changes.
- `go test` may fail if vet warnings are introduced; treat vet-clean output as required.

## Change Guidelines

1. Keep package responsibilities separated (prompt/config/program/export/hevy).
2. Avoid introducing new dependencies unless explicitly justified.
3. Prefer small, deterministic functions for set-generation and conversion logic.
4. When changing Hevy API payloads, keep wrapper shapes intact (`{"routine": ...}`, `{"routine_folder": ...}`).
5. Do not hardcode or log API keys; they come from interactive input.
6. Keep CSV schema stable unless requested (`Week, Day, Exercise, Sets, Reps, Weight, Percentage`).

## Testing Priorities

When adding/modifying logic, prioritize tests for:

1. `config`: training max and rounding behavior.
2. `program`: week schemes, deload behavior, BBB/accessory inclusion, day ordering.
3. `hevy/converter`: AMRAP translation to `rep_range`, warmup set typing, lbs->kg conversion.
4. `hevy/client`: response decoding and error handling using `httptest` or mock transport.

## Safety

- Never commit secrets (Hevy API keys, tokens, personal exports).
- Keep generated artifacts (`*.csv`, binaries) out of commits unless explicitly requested.
