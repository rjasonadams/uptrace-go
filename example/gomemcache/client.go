package main

import (
	"context"
	"errors"
	"flag"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/contrib/instrumentation/github.com/bradfitz/gomemcache/memcache/otelmemcache"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("memcache-tracer")

func main() {
	flag.Parse()

	ctx := context.Background()

	upclient := uptrace.NewClient(&uptrace.Config{
		// copy your project DSN here or use UPTRACE_DSN enar
		DSN: "",
	})
	defer upclient.Close()
	defer upclient.ReportPanic(ctx)

	upclient.ReportError(ctx, errors.New("hello from uptrace-go!"))

	mc := otelmemcache.NewClientWithTracing(
		memcache.New("memcached-server:11211"),
	)

	ctx, s := tracer.Start(ctx, "test-operations")
	doMemcacheOperations(ctx, mc)
	s.End()
}

func doMemcacheOperations(ctx context.Context, mc *otelmemcache.Client) {
	mc = mc.WithContext(ctx)

	err := mc.Add(&memcache.Item{
		Key:   "foo",
		Value: []byte("bar"),
	})
	if err != nil {
		trace.SpanFromContext(ctx).RecordError(err)
	}

	_, err = mc.Get("foo")
	if err != nil {
		trace.SpanFromContext(ctx).RecordError(err)
	}

	_, err = mc.Get("hello")
	if err != nil {
		trace.SpanFromContext(ctx).RecordError(err)
	}

	err = mc.Delete("foo")
	if err != nil {
		trace.SpanFromContext(ctx).RecordError(err)
	}
}
