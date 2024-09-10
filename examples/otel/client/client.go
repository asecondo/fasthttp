package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/valyala/fasthttp"
	fasthttpotel "github.com/valyala/fasthttp/telemetry/otel"
	"go.opentelemetry.io/otel"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	tp, err := initTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()
	// url := flag.String("server", "http://localhost:7777/hello", "server url")
	// flag.Parse()
	url := "http://localhost:7777/hello"
	client := fasthttp.HostClient{
		Addr:      "localhost:7777",
		Transport: fasthttpotel.NewTransport(fasthttp.DefaultTransport),
	}

	tracer := otel.Tracer("example/client")
	err = func(ctx context.Context) error {
		ctx, span := tracer.Start(context.Background(), "ping")
		defer span.End()
		fastUrl := fasthttp.AcquireURI()
		err := fastUrl.Parse(nil, []byte(url))
		if err != nil {
			fmt.Printf("Error parsing url: %v\n", err)
			return err
		}
		req := fasthttp.AcquireRequest()
		req.SetURI(fastUrl)
		fasthttp.ReleaseURI(fastUrl)

		req.Header.SetMethod(fasthttp.MethodGet)
		// Need to explicitly set context to propagate the parent span
		req.SetContext(ctx)
		resp := fasthttp.AcquireResponse()
		fmt.Println(req.URI())
		err = client.Do(req, resp)
		fasthttp.ReleaseRequest(req)
		if err != nil {
			fmt.Printf("Error from client: %v\n", err)
		} else {
			fmt.Printf("Response: %s\n", resp.Body())
		}
		fasthttp.ReleaseResponse(resp)

		return nil
	}(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("Response Received: %s\n\n\n", body)
	fmt.Printf("Waiting for few seconds to export spans ...\n\n")
	time.Sleep(10 * time.Second)
	fmt.Printf("Inspect traces on stdout\n")
}

func initTracer() (*sdktrace.TracerProvider, error) {
	// Create stdout exporter to be able to retrieve
	// the collected spans.
	exporter, err := stdout.New(stdout.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, err
}
