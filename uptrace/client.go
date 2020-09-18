package uptrace

import (
	"context"
	"runtime"

	"github.com/uptrace/uptrace-go/internal"
	"github.com/uptrace/uptrace-go/spanexp"
	"github.com/uptrace/uptrace-go/upconfig"

	"go.opentelemetry.io/otel/api/global"
	apitrace "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const dummySpanName = "__dummy__"

type Config = upconfig.Config

// Client represents Uptrace client.
type Client struct {
	cfg *Config

	tracer apitrace.Tracer

	sp       sdktrace.SpanProcessor
	provider *sdktrace.Provider
}

func NewClient(cfg *Config) *Client {
	cfg.Init()

	client := &Client{
		cfg: cfg,

		tracer: global.Tracer("github.com/uptrace/uptrace-go"),
	}
	client.setupTracing()

	return client
}

// Closes closes the client releasing associated resources.
func (c *Client) Close() error {
	runtime.Gosched()
	c.provider.UnregisterSpanProcessor(c.sp)
	return nil
}

// ReportError reports an error as a span event creating a dummy span if necessary.
func (c *Client) ReportError(ctx context.Context, err error, opts ...apitrace.ErrorOption) {
	span := apitrace.SpanFromContext(ctx)
	if !span.IsRecording() {
		ctx, span = c.tracer.Start(ctx, dummySpanName)
		defer span.End()
	}

	span.RecordError(ctx, err, opts...)
}

// ReportPanic is used with defer to report panics.
func (c *Client) ReportPanic(ctx context.Context) {
	val := recover()
	if val == nil {
		return
	}

	span := apitrace.SpanFromContext(ctx)
	isRecording := span.IsRecording()
	if !isRecording {
		ctx, span = c.tracer.Start(ctx, dummySpanName)
	}

	span.AddEvent(
		ctx,
		"log",
		label.String("log.severity", "panic"),
		label.Any("log.message", val),
	)

	if !isRecording {
		// Can't use `defer span.End()` because it recovers from panic too.
		span.End()
	}

	// Re-throw the panic.
	panic(val)
}

//------------------------------------------------------------------------------

// Tracer returns a named tracer that exports span to Uptrace.
func (c *Client) Tracer(name string) apitrace.Tracer {
	return global.Tracer(name)
}

func (c *Client) setupTracing() {
	if c.cfg.Disabled {
		return
	}

	var err error

	kvs := make([]label.KeyValue, 0, len(c.cfg.Resource))
	for k, v := range c.cfg.Resource {
		kvs = append(kvs, label.Any(k, v))
	}

	var res *resource.Resource
	if len(kvs) > 0 {
		res = resource.New(kvs...)
	}

	c.provider, err = sdktrace.NewProvider(
		sdktrace.WithConfig(sdktrace.Config{
			Resource:       res,
			DefaultSampler: c.cfg.Sampler,
		}),
	)
	if err != nil {
		internal.Logger.Printf("NewProvider failed: %s", err)
		return
	}

	c.sp, err = spanexp.NewBatchSpanProcessor(c.cfg)
	if err != nil {
		internal.Logger.Printf("NewBatchSpanProcessor failed: %s", err)
		return
	}

	c.provider.RegisterSpanProcessor(c.sp)
	global.SetTraceProvider(c.provider)
}
