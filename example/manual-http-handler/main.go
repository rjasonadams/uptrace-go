package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/semconv"
)

var tracer = otel.Tracer("github.com/your/repo")

func main() {
	ctx := context.Background()

	upclient := uptrace.NewClient(&uptrace.Config{
		// copy your project DSN here or use UPTRACE_DSN env var
		DSN: "",
	})
	defer upclient.Close()
	defer upclient.ReportPanic(ctx)

	// HTTP handler span.
	_, span := tracer.Start(ctx, "GET /articles/:articleID")
	defer span.End()

	// Required attributes.
	span.SetAttributes(
		semconv.HTTPMethodKey.String("GET"),
		semconv.HTTPRouteKey.String("/articles/:articleID"),
	)

	// Optional attributes.
	span.SetAttributes(
		semconv.HTTPTargetKey.String("/articles/123"),
		semconv.HTTPStatusCodeKey.Int(http.StatusOK),
		semconv.HTTPUserAgentKey.String("Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0) Gecko/20100101 Firefox/47.0"),
		semconv.HTTPClientIPKey.String("127.0.0.1"),

		label.String("code.function", "articleEndpoint"),
		label.String("code.filepath", "/var/lib/site/article/article_api.go"),
		label.Int("code.lineno", 55),
	)

	fmt.Printf("trace: %s\n", upclient.TraceURL(span))
}
