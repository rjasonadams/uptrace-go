package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/uptrace/uptrace-go/extra/otellogrus"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel/api/global"
)

func main() {
	ctx := context.Background()
	upclient := setupUptrace()

	defer upclient.Close()
	defer upclient.ReportPanic(ctx)

	// Add OpenTelemetry logging hook.
	logrus.AddHook(otellogrus.NewLoggingHook())

	tracer := global.Tracer("example")

	ctx, span := tracer.Start(ctx, "main")
	defer span.End()

	logrus.WithContext(ctx).Error("hello")

	fmt.Printf("trace: %s\n", upclient.TraceURL(span))
}

func setupUptrace() *uptrace.Client {
	if os.Getenv("UPTRACE_DSN") == "" {
		panic("UPTRACE_DSN is empty or missing")
	}

	upclient := uptrace.NewClient(&uptrace.Config{
		// copy your project DSN here or use UPTRACE_DSN env var
		DSN: "",
	})

	return upclient
}
