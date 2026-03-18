# go-cron-kit

[![CI](https://github.com/philiprehberger/go-cron-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/philiprehberger/go-cron-kit/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/philiprehberger/go-cron-kit.svg)](https://pkg.go.dev/github.com/philiprehberger/go-cron-kit)
[![License](https://img.shields.io/github/license/philiprehberger/go-cron-kit)](LICENSE)

Cron expression parser and job scheduler with overlap prevention for Go

## Installation

```bash
go get github.com/philiprehberger/go-cron-kit
```

## Usage

### Parse Cron Expressions

```go
import "github.com/philiprehberger/go-cron-kit"

sched, err := cronkit.Parse("*/5 * * * *") // every 5 minutes
next := sched.Next(time.Now())
fmt.Println(next)
```

### Scheduler

```go
s := cronkit.NewScheduler()

s.Add("cleanup", "0 2 * * *", func(ctx context.Context) {
    // runs daily at 2 AM
    cleanupOldRecords(ctx)
})

s.Add("heartbeat", "*/5 * * * *", func(ctx context.Context) {
    // runs every 5 minutes
    sendHeartbeat(ctx)
})

// Start blocks until context is cancelled
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
go s.Start(ctx)
```

### Overlap Prevention

Jobs that are still running when the next tick fires are automatically skipped.

### Next Run Preview

```go
next, ok := s.NextRun("cleanup")
if ok {
    fmt.Printf("Next cleanup: %s\n", next)
}
```

### Job Management

```go
names := s.Jobs() // ["cleanup", "heartbeat"]
s.Stop()          // graceful shutdown
```

### Supported Cron Syntax

Standard 5-field format: `minute hour day-of-month month day-of-week`

- `*` — any value
- `*/n` — every n
- `1-5` — ranges
- `1,3,5` — lists

When both day-of-month and day-of-week are restricted (not `*`), the job runs when **either** matches (POSIX cron OR semantics). For example, `0 0 15 * 1` runs at midnight on the 15th **or** on Mondays.

## API

| Function / Method | Description |
|---|---|
| `Parse(expr string) (*Schedule, error)` | Parse a standard 5-field cron expression |
| `NewScheduler() *Scheduler` | Create a new job scheduler |
| `Schedule.Next(after time.Time) time.Time` | Return the next time after t that matches the schedule |
| `Scheduler.Add(name, expr string, handler func(ctx context.Context)) error` | Register a new job with the given cron expression |
| `Scheduler.Start(ctx context.Context)` | Start the scheduler, blocking until cancelled or stopped |
| `Scheduler.Stop()` | Signal the scheduler to stop gracefully |
| `Scheduler.NextRun(name string) (time.Time, bool)` | Return the next scheduled run time for a named job |
| `Scheduler.Jobs() []string` | Return a copy of all registered job names |

**Types:**

| Type | Description |
|---|---|
| `Schedule` | A parsed cron expression with Minutes, Hours, DaysOfMonth, Months, and DaysOfWeek fields |
| `Job` | A scheduled job with Name, Schedule, and Handler fields |
| `Scheduler` | Manages cron jobs with overlap prevention |

## Development

```bash
go test ./...
go vet ./...
```

## License

MIT
