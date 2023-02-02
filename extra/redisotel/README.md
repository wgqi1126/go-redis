# OpenTelemetry instrumentation for go-redis

## Installation

```bash
go get github.com/wgqi1126/go-redis/extra/redisotel/v9
```

## Usage

Tracing is enabled by adding a hook:

```go
import (
    "github.com/wgqi1126/go-redis/v9"
    "github.com/wgqi1126/go-redis/extra/redisotel/v9"
)

rdb := rdb.NewClient(&rdb.Options{...})

rdb.AddHook(redisotel.NewTracingHook())
```

See [example](example) and
[Monitoring Go Redis Performance and Errors](https://redis.uptrace.dev/guide/go-redis-monitoring.html)
for details.
