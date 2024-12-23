package otel

import (
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	scopeName = "github.com/valyala/fasthttp/telemetry/otel"
)

type config struct {
	Propagators       propagation.TextMapPropagator
	SpanNameFormatter func(*fasthttp.Request) string
	SpanStartOptions  []trace.SpanStartOption
	Tracer            trace.Tracer
	TracerProvider    trace.TracerProvider
}

// Option sets the tracing option values.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

func newConfig(opts ...Option) *config {
	c := &config{
		Propagators: otel.GetTextMapPropagator(),
	}

	for _, opt := range opts {
		opt.apply(c)
	}

	if c.TracerProvider != nil {
		c.Tracer = newTracer(c.TracerProvider)
	}

	return c
}

// WithTracerProvider sets the tracer provider to use.
func WithTracerProvider(provider trace.TracerProvider) Option {
	return optionFunc(func(cfg *config) {
		if provider != nil {
			cfg.TracerProvider = provider
		}
	})
}

// WithPropagators sets the propagators to use.
func WithPropagators(propagators propagation.TextMapPropagator) Option {
	return optionFunc(func(cfg *config) {
		if propagators != nil {
			cfg.Propagators = propagators
		}
	})
}

func WithSpanNameFormatter(formatter func(*fasthttp.Request) string) Option {
	return optionFunc(func(cfg *config) {
		cfg.SpanNameFormatter = formatter
	})
}

func WithSpanStartOptions(opts ...trace.SpanStartOption) Option {
	return optionFunc(func(cfg *config) {
		cfg.SpanStartOptions = opts
	})
}

func newTracer(tp trace.TracerProvider) trace.Tracer {
	return tp.Tracer(scopeName)
}
