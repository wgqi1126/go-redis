package redisotel

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/wgqi1126/go-redis/extra/rediscmd/v9"
	"github.com/wgqi1126/go-redis/v9"
)

const (
	instrumName = "github.com/wgqi1126/go-redis/extra/redisotel"
)

func InstrumentTracing(rdb redis.UniversalClient, opts ...TracingOption) error {
	switch rdb := rdb.(type) {
	case *redis.Client:
		opt := rdb.Options()
		connString := formatDBConnString(opt.Network, opt.Addr)
		rdb.AddHook(newTracingHook(connString, opts...))
		return nil
	case *redis.ClusterClient:
		rdb.AddHook(newTracingHook("", opts...))

		rdb.OnNewNode(func(rdb *redis.Client) {
			opt := rdb.Options()
			connString := formatDBConnString(opt.Network, opt.Addr)
			rdb.AddHook(newTracingHook(connString, opts...))
		})
		return nil
	case *redis.Ring:
		rdb.AddHook(newTracingHook("", opts...))

		rdb.OnNewNode(func(rdb *redis.Client) {
			opt := rdb.Options()
			connString := formatDBConnString(opt.Network, opt.Addr)
			rdb.AddHook(newTracingHook(connString, opts...))
		})
		return nil
	default:
		return fmt.Errorf("redisotel: %T not supported", rdb)
	}
}

type tracingHook struct {
	conf *config

	spanOpts []trace.SpanStartOption
}

var _ redis.Hook = (*tracingHook)(nil)

func newTracingHook(connString string, opts ...TracingOption) *tracingHook {
	baseOpts := make([]baseOption, len(opts))
	for i, opt := range opts {
		baseOpts[i] = opt
	}
	conf := newConfig(baseOpts...)

	if conf.tracer == nil {
		conf.tracer = conf.tp.Tracer(
			instrumName,
			trace.WithInstrumentationVersion("semver:"+redis.Version()),
		)
	}
	if connString != "" {
		conf.attrs = append(conf.attrs, semconv.DBConnectionStringKey.String(connString))
	}

	return &tracingHook{
		conf: conf,

		spanOpts: []trace.SpanStartOption{
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(conf.attrs...),
		},
	}
}

func (th *tracingHook) DialHook(hook redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		if !trace.SpanFromContext(ctx).IsRecording() {
			return hook(ctx, network, addr)
		}

		ctx, span := th.conf.tracer.Start(ctx, "redis.dial", th.spanOpts...)
		defer span.End()

		conn, err := hook(ctx, network, addr)
		if err != nil {
			recordError(span, err)
			return nil, err
		}
		return conn, nil
	}
}

func (th *tracingHook) ProcessHook(hook redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		if !trace.SpanFromContext(ctx).IsRecording() {
			return hook(ctx, cmd)
		}

		fn, file, line := funcFileLine("github.com/wgqi1126/go-redis")

		attrs := make([]attribute.KeyValue, 0, 8)
		attrs = append(attrs,
			semconv.CodeFunctionKey.String(fn),
			semconv.CodeFilepathKey.String(file),
			semconv.CodeLineNumberKey.Int(line),
		)

		if th.conf.dbStmtEnabled {
			cmdString := rediscmd.CmdString(cmd)
			attrs = append(attrs, semconv.DBStatementKey.String(cmdString))
		}

		opts := th.spanOpts
		opts = append(opts, trace.WithAttributes(attrs...))

		ctx, span := th.conf.tracer.Start(ctx, cmd.FullName(), opts...)
		defer span.End()

		if err := hook(ctx, cmd); err != nil {
			recordError(span, err)
			return err
		}
		return nil
	}
}

func (th *tracingHook) ProcessPipelineHook(
	hook redis.ProcessPipelineHook,
) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		if !trace.SpanFromContext(ctx).IsRecording() {
			return hook(ctx, cmds)
		}

		fn, file, line := funcFileLine("github.com/wgqi1126/go-redis")

		attrs := make([]attribute.KeyValue, 0, 8)
		attrs = append(attrs,
			semconv.CodeFunctionKey.String(fn),
			semconv.CodeFilepathKey.String(file),
			semconv.CodeLineNumberKey.Int(line),
			attribute.Int("db.redis.num_cmd", len(cmds)),
		)

		summary, cmdsString := rediscmd.CmdsString(cmds)
		if th.conf.dbStmtEnabled {
			attrs = append(attrs, semconv.DBStatementKey.String(cmdsString))
		}

		opts := th.spanOpts
		opts = append(opts, trace.WithAttributes(attrs...))

		ctx, span := th.conf.tracer.Start(ctx, "redis.pipeline "+summary, opts...)
		defer span.End()

		if err := hook(ctx, cmds); err != nil {
			recordError(span, err)
			return err
		}
		return nil
	}
}

func recordError(span trace.Span, err error) {
	if err != redis.Nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

func formatDBConnString(network, addr string) string {
	if network == "tcp" {
		network = "redis"
	}
	return fmt.Sprintf("%s://%s", network, addr)
}

func funcFileLine(pkg string) (string, string, int) {
	const depth = 16
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	ff := runtime.CallersFrames(pcs[:n])

	var fn, file string
	var line int
	for {
		f, ok := ff.Next()
		if !ok {
			break
		}
		fn, file, line = f.Function, f.File, f.Line
		if !strings.Contains(fn, pkg) {
			break
		}
	}

	if ind := strings.LastIndexByte(fn, '/'); ind != -1 {
		fn = fn[ind+1:]
	}

	return fn, file, line
}
