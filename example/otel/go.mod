module github.com/wgqi1126/go-redis/example/otel

go 1.14

replace github.com/wgqi1126/go-redis/v9 => ../..

replace github.com/wgqi1126/go-redis/extra/redisotel/v9 => ../../extra/redisotel

replace github.com/wgqi1126/go-redis/extra/rediscmd/v9 => ../../extra/rediscmd

require (
	github.com/wgqi1126/go-redis/extra/redisotel/v9 v9.0.2
	github.com/wgqi1126/go-redis/v9 v9.0.2
	github.com/uptrace/uptrace-go v1.12.0
	go.opentelemetry.io/otel v1.12.0
	google.golang.org/genproto v0.0.0-20230131230820-1c016267d619 // indirect
)
