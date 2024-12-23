package otel

import (
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Transport struct {
	propagators       propagation.TextMapPropagator
	roundTripper      fasthttp.RoundTripper
	spanNameFormatter func(*fasthttp.Request) string
	spanStartOptions  []trace.SpanStartOption
	tracer            trace.Tracer
}

func NewTransport(base fasthttp.RoundTripper, opts ...Option) *Transport {
	if base == nil {
		base = fasthttp.DefaultTransport
	}

	transport := Transport{
		roundTripper: base,
	}

	defaultOpts := []Option{
		WithSpanStartOptions(trace.WithSpanKind(trace.SpanKindClient)),
		WithSpanNameFormatter(defaultTransportFormatter),
	}

	config := newConfig(append(defaultOpts, opts...)...)
	transport.applyConfig(config)

	return &transport
}

func (t *Transport) applyConfig(c *config) {
	t.propagators = c.Propagators
	t.spanNameFormatter = c.SpanNameFormatter
	t.spanStartOptions = c.SpanStartOptions
	t.tracer = c.Tracer
}

func defaultTransportFormatter(req *fasthttp.Request) string {
	// Avoid returning '"HTTP " + string(req.Header.Method())' to avoid string allocations.
	switch b2s(req.Header.Method()) {
	case "GET":
		return "HTTP GET"
	case "POST":
		return "HTTP POST"
	case "PUT":
		return "HTTP PUT"
	case "DELETE":
		return "HTTP DELETE"
	case "PATCH":
		return "HTTP PATCH"
	case "HEAD":
		return "HTTP HEAD"
	case "OPTIONS":
		return "HTTP OPTIONS"
	case "CONNECT":
		return "HTTP CONNECT"
	case "TRACE":
		return "HTTP TRACE"
	default:
		return "HTTP"
	}
}

func (t *Transport) RoundTrip(
	hc *fasthttp.HostClient,
	req *fasthttp.Request,
	resp *fasthttp.Response,
) (retry bool, err error) {
	// TODO (NOW): consider adding filters for the MVP. if any of the filters apply just passthrough
	//  to the base roundTripper.

	tracer := t.tracer
	if tracer == nil {
		if span := trace.SpanFromContext(req.Context()); span.SpanContext().IsValid() {
			tracer = newTracer(span.TracerProvider())
		} else {
			tracer = newTracer(otel.GetTracerProvider())
		}
	}

	ctx, span := tracer.Start(req.Context(), defaultTransportFormatter(req), t.spanStartOptions...)
	defer span.End()

	carrier := &headerCarrier{
		header: &req.Header,
	}
	t.propagators.Inject(ctx, carrier)

	setRequestAttributes(span, req)

	retry, err = t.roundTripper.RoundTrip(hc, req, resp)

	setResponseAttributes(span, resp)
	if err != nil {
		// TODO (NOW): we technically need to define and record an error.type attribute if applicable.
		//  for now, we'll just record the error.
		span.RecordError(err)
	}

	return retry, err
}

func setRequestAttributes(span trace.Span, req *fasthttp.Request) {
	// TODO (NOW): deal with max number of kvs generated for requests
	attrs := make([]attribute.KeyValue, 0)
	attrs = append(
		attrs,
		attribute.String("http.request.method", b2s(req.Header.Method())),
		attribute.String("user_agent.original", b2s(req.Header.UserAgent())),
		attribute.String("server.address", b2s(req.Host())),
		// TODO (NOW): port isn't explicitly stored. need to parse it from host field, but going to skip for now.
		// attribute.String("server.port", string(req.Host())),
		attribute.String("url.full", b2s(req.URI().FullURI())),
		attribute.String("url.scheme", b2s(req.URI().Scheme())),
		attribute.String("network.transport", "tcp"),
		attribute.String("network.protocol.version", b2s(req.Header.Protocol())),
	)
	span.SetAttributes(attrs...)
}

func setResponseAttributes(span trace.Span, resp *fasthttp.Response) {
	// TODO (NOW): deal with max number of kvs generated for responses
	attrs := make([]attribute.KeyValue, 0)
	statusCode := resp.StatusCode()
	if statusCode != 0 {
		attrs = append(attrs, attribute.Int("http.response.status_code", resp.StatusCode()))
	}
	if statusCode < 100 || statusCode >= 600 {
		span.SetStatus(codes.Error, "Invalid HTTP status code")
	} else if statusCode >= 400 {
		span.SetStatus(codes.Error, "")
	}
	attrs = append(attrs, attribute.Int("http.response.content_length", len(resp.Body())))
	span.SetAttributes(attrs...)
}
