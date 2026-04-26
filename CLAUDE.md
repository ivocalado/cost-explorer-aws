# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

A Go CLI that queries the AWS Cost Explorer API (`aws-sdk-go-v2`) to report on AWS costs: last month total, daily-by-service, last-N-days, tag-filtered costs, and current-month forecasts. User-facing output is in Portuguese (pt-BR).

Module name is `example` (Go 1.24.4). The `ce/` directory holds the package `mycostexplorer`, imported as `example/ce`.

## Commands

```bash
go build                          # produce ./example binary
go run . <subcommand> [flags]     # run without building
go run . help                     # list all subcommands
go vet ./...
go mod tidy
```

Every subcommand accepts `-profile` (AWS shared-config profile, optional) and `-region` (default `us-east-1` — Cost Explorer is a global service hosted there). Run `go run . <subcommand> -h` for per-command flags.

There is no test suite; no `*_test.go` files exist.

## Architecture

**Two-layer dispatch.** `main.go` is a flat `switch` over `os.Args[1]` — each case builds its own `flag.FlagSet`, constructs a `*costexplorer.Client` via `newCostExplorerClient`, and calls one exported function from `ce/`. Each file in `ce/` implements one or two related subcommands.

**Adding a new subcommand requires three coordinated edits:** (1) implement `GetXxx(ctx, ce, ...) error` in a new `ce/*.go` file, (2) add a `case` in `main.go`'s switch with its own `FlagSet`, (3) add a line to the `usage()` help text. Missing any of these leaves the command undiscoverable or unreachable.

**Shared helpers live in sibling files, not a separate utils file:**
- `lastClosedMonthRange()` — defined in `ce/last_month.go`
- `currentMonthRange()` and `lastNDaysRange(n)` — defined in `ce/last_n_days.go`

These are package-private and reused across files (e.g. `cost_by_tag.go` calls helpers from `last_n_days.go`). When editing date logic, check all callers.

## Conventions to preserve

- **Metric is always `UnblendedCost`** — hardcoded as a local string in each function, not a package constant.
- **Date format is `"2006-01-02"`, all times in UTC.** `End` of `DateInterval` is exclusive (e.g. `lastNDaysRange` returns `tomorrow` as end to include today).
- **Amounts are parsed from strings via `fmt.Sscanf(s, "%f", &cost)`.** The SDK returns `*string` amounts; the codebase uses `aws.ToString(...)` to deref then `Sscanf` to convert. Costs `<= 0.01` are filtered out of per-line output but still summed into totals.
- **Output is Portuguese with emoji/box-drawing decoration** (`📅`, `═══`, `───`). Keep new commands stylistically consistent — column widths use `%-40s` for service names and `%10.2f USD` for amounts.
- **Errors bubble up wrapped:** `return fmt.Errorf("GetCostAndUsage: %w", err)`. `main.go` calls `log.Fatalf` on error.
- **Forecast quirk:** `GetCostForecast` only accepts a `Start` of today or future. `forecast.go` uses `now.Format(layout)` as the start and subtracts month-to-date cost from the forecast total to derive the "remaining" estimate — don't try to forecast a past period.
