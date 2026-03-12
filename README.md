# go-cron-kit

Cron expression parser and job scheduler with overlap prevention for Go.

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

## License

MIT
