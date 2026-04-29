# cost-explorer-aws

A small Go CLI built on top of the AWS Cost Explorer API (`aws-sdk-go-v2`) to
inspect, break down, and forecast AWS spend straight from the terminal — no
console clicking, no dashboard. User-facing output is in Portuguese (pt-BR).

## Purpose

AWS bills are easy to read at the end of the month and hard to read at any
other moment. This tool gives you a fast, scriptable view of how much you are
spending, where it is going, and where it is heading, so cost questions can be
answered in seconds without leaving the shell.

It is intentionally minimal: one binary, one subcommand per question, output
designed to be readable in a terminal or piped into a file.

## Information you can obtain

Each subcommand answers one specific question:

| Command | What it reports |
|---|---|
| `last-month-total` | Total `UnblendedCost` for the last fully closed month |
| `daily-by-service` | Daily cost broken down by AWS service for the last closed month |
| `last-n-daily-by-service` | Daily cost by service over the last N days (or current month if N is omitted) |
| `cost-by-tag` | Total cost over the last N days filtered by a `tag-key`/`tag-value` |
| `cost-by-tag-detailed` | Daily cost by service over the last N days filtered by a tag |
| `forecast` | Current-month forecast vs. month-to-date and previous month, with a variation alert |
| `forecast-by-service` | Current-month forecast plus accumulated cost grouped by service |

All amounts are in USD using the `UnblendedCost` metric. Line items below
`$0.01` are hidden but still summed into totals.

Common flags:

- `-profile` — AWS shared-config profile (optional; uses the default chain if omitted).
- `-region` — defaults to `us-east-1` (Cost Explorer is a global service hosted there).
- `-days` — used by the `last-n-*` and `cost-by-tag*` commands.
- `-tag-key` / `-tag-value` — required for the tag-filtered commands.

Run `go run . <command> -h` for the per-command flag list.

## Use case scenarios

- **Daily cost check.** Run `daily-by-service` or `last-n-daily-by-service -days 7`
  every morning to spot the service that quietly doubled overnight before it
  shows up on the invoice.
- **End-of-month closeout.** `last-month-total` gives the single number you
  need for a finance report or a Slack update; pipe it to a file and attach.
- **Per-project / per-team chargeback.** Tag your resources with
  `Project=<name>` or `Team=<name>` and use `cost-by-tag -tag-key Project -tag-value billing -days 30`
  to bill internal stakeholders. Use `cost-by-tag-detailed` when someone asks
  *which* service inside that project is driving the cost.
- **Environment isolation.** Compare `Environment=Production` vs.
  `Environment=Staging` using `cost-by-tag` to make sure non-prod isn't
  silently outspending prod.
- **Mid-month budget guardrail.** `forecast` shows month-to-date, projected
  remainder, and percent variation against last month, with a built-in alert
  when the projected increase passes 5% / 10%. Drop it in a cron job and
  forward to Slack/email when the alert fires.
- **Spotting the outlier service.** `forecast-by-service` accumulates current
  spend per service and shows the projected total — useful to catch a
  misconfigured service (NAT Gateway, CloudWatch Logs, data transfer) before
  it dominates the bill.
- **Cost spike investigation.** When the forecast jumps, run
  `last-n-daily-by-service -days 14` to see which service started the climb
  and on which day.

## Quick start

```bash
go build                          # produces ./example
go run . help                     # list subcommands
go run . last-month-total
go run . forecast -profile prod
go run . cost-by-tag -tag-key Project -tag-value billing -days 30
```

The caller's IAM principal needs `ce:GetCostAndUsage` and `ce:GetCostForecast`.

## Project layout

- `main.go` — flat `switch` dispatcher; each subcommand owns its own `flag.FlagSet`.
- `ce/` — package `mycostexplorer` (imported as `example/ce`); one file per
  feature area (`last_month.go`, `daily_by_service.go`, `last_n_days.go`,
  `cost_by_tag.go`, `forecast.go`).

See `CLAUDE.md` for conventions to preserve when extending the tool (date
handling, output style, error wrapping, forecast quirks).
