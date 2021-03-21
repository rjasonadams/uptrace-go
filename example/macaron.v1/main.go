package main

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/instrumentation/gopkg.in/macaron.v1/otelmacaron"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/macaron.v1"

	"github.com/uptrace/uptrace-go/uptrace"
)

var tracer = otel.Tracer("app_or_package_name")

func main() {
	ctx := context.Background()

	uptrace.ConfigureOpentelemetry(&uptrace.Config{
		// copy your project DSN here or use UPTRACE_DSN env var
		DSN: "",
	})
	defer uptrace.Shutdown(ctx)

	m := macaron.Classic()
	m.Get("/profiles/:username", userProfileEndpoint)
	m.Use(otelmacaron.Middleware("service-name"))

	m.Run(9999)
}

func userProfileEndpoint(c *macaron.Context) string {
	ctx := c.Req.Context()

	username := c.Params("username")
	name, err := selectUser(ctx, username)
	if err != nil {
		trace.SpanFromContext(ctx).RecordError(err)
	}

	return fmt.Sprintf(`<html><h1>Hello %s %s </h1></html>`+"\n", username, name)
}

func selectUser(ctx context.Context, username string) (string, error) {
	_, span := tracer.Start(ctx, "selectUser")
	defer span.End()

	span.SetAttributes(attribute.String("username", username))

	if username == "admin" {
		return "Joe", nil
	}

	return "", fmt.Errorf("username=%s not found", username)
}