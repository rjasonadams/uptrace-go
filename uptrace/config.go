package uptrace

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type config struct {
	DSN string

	// Common options

	ServiceName        string
	ServiceVersion     string
	ResourceAttributes []attribute.KeyValue
	Resource           *resource.Resource

	// Tracing options

	TracingDisabled   bool
	TextMapPropagator propagation.TextMapPropagator
	TracerProvider    *sdktrace.TracerProvider
	TraceSampler      sdktrace.Sampler
	PrettyPrint       bool

	// Metrics options

	MetricsDisabled bool
}

func newConfig(opts []Option) *config {
	cfg := new(config)

	if dsn, ok := os.LookupEnv("UPTRACE_DSN"); ok {
		cfg.DSN = dsn
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return cfg
}

func (cfg *config) newResource() *resource.Resource {
	if cfg.Resource != nil {
		return cfg.Resource
	}
	return buildResource(cfg.ResourceAttributes, cfg.ServiceName, cfg.ServiceVersion)
}

func buildResource(
	resourceAttributes []attribute.KeyValue,
	serviceName, serviceVersion string,
) *resource.Resource {
	ctx := context.TODO()

	var attrs []attribute.KeyValue
	attrs = append(attrs, resourceAttributes...)
	if serviceName != "" {
		attrs = append(attrs, semconv.ServiceNameKey.String(serviceName))
	}
	if serviceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersionKey.String(serviceVersion))
	}

	res, _ := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(attrs...))
	if res == nil {
		return resource.Environment()
	}
	return res
}

//------------------------------------------------------------------------------

type Option interface {
	apply(cfg *config)
}

type option func(cfg *config)

func (fn option) apply(cfg *config) {
	fn(cfg)
}

// WithDSN configures a data source name that is used to connect to Uptrace, for example,
// `https://<token>@api.uptrace.dev/<project_id>`.
//
// The default is to use UPTRACE_DSN environment variable.
func WithDSN(dsn string) Option {
	return option(func(cfg *config) {
		cfg.DSN = dsn
	})
}

// WithServiceVersion configures a `service.name` resource attribute.
func WithServiceName(serviceName string) Option {
	return option(func(cfg *config) {
		cfg.ServiceName = serviceName
	})
}

// WithServiceVersion configures a `service.version` resource attribute, for example, `1.0.0`.
func WithServiceVersion(serviceVersion string) Option {
	return option(func(cfg *config) {
		cfg.ServiceVersion = serviceVersion
	})
}

// WithResourceAttributes configures resource attributes that describe an entity that produces
// telemetry, for example, such attributes as host.name, service.name, etc.
//
// The default is to use `OTEL_RESOURCE_ATTRIBUTES` env var, for example,
// `OTEL_RESOURCE_ATTRIBUTES=service.name=myservice,service.version=1.0.0`.
func WithResourceAttributes(resourceAttributes []attribute.KeyValue) Option {
	return option(func(cfg *config) {
		cfg.ResourceAttributes = resourceAttributes
	})
}

// WithResource configures a resource that describes an entity that produces telemetry,
// for example, such attributes as host.name and service.name. All produced spans and metrics
// will have these attributes.
//
// WithResource overrides and replaces any other resource attributes.
func WithResource(resource *resource.Resource) Option {
	return option(func(cfg *config) {
		cfg.Resource = resource
	})
}

//------------------------------------------------------------------------------

type TracingOption interface {
	Option
	tracing()
}

type tracingOption func(cfg *config)

var _ TracingOption = (*tracingOption)(nil)

func (fn tracingOption) apply(cfg *config) {
	fn(cfg)
}

func (fn tracingOption) tracing() {}

// TracingDisabled can be used to skip tracing configuration.
func WithTracingDisabled() TracingOption {
	return tracingOption(func(cfg *config) {
		cfg.TracingDisabled = true
	})
}

// TracerProvider overwrites the default Uptrace tracer provider.
// You can use it to configure Uptrace distro to use OTLP exporter.
func WithTracerProvider(provider *sdktrace.TracerProvider) TracingOption {
	return tracingOption(func(cfg *config) {
		cfg.TracerProvider = provider
	})
}

// WithTraceSampler configures a span sampler.
func WithTraceSampler(sampler sdktrace.Sampler) TracingOption {
	return tracingOption(func(cfg *config) {
		cfg.TraceSampler = sampler
	})
}

// WithTextMapPropagator sets the global TextMapPropagator used by OpenTelemetry.
// The default is propagation.TraceContext and propagation.Baggage.
func WithTextMapPropagator(propagator propagation.TextMapPropagator) TracingOption {
	return tracingOption(func(cfg *config) {
		cfg.TextMapPropagator = propagator
	})
}

// WithPrettyPrintSpanExporter adds a span exproter that prints spans to stdout.
// It is useful for debugging or demonstration purposes.
func WithPrettyPrintSpanExporter() TracingOption {
	return tracingOption(func(cfg *config) {
		cfg.PrettyPrint = true
	})
}

//------------------------------------------------------------------------------

type MetricsOption interface {
	Option
	metrics()
}

type metricsOption func(cfg *config)

var _ MetricsOption = (*metricsOption)(nil)

func (fn metricsOption) apply(cfg *config) {
	fn(cfg)
}

func (fn metricsOption) metrics() {}

// WithMetricsDisabled can be used to skip metrics configuration.
func WithMetricsDisabled() MetricsOption {
	return metricsOption(func(cfg *config) {
		cfg.MetricsDisabled = true
	})
}
