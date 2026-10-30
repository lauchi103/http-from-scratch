package main

import (
	"context"
	"fmt"
	"httpServer/internal/http1"
	"httpServer/internal/mymetrics"
	"io"
	"log"
	"net/http"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const addr = ":8080"

var tracer trace.Tracer

func newExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "vts-vt-single-server.monitoring.svc.cluster.local:10428"
	}
	return otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithURLPath("/insert/opentelemetry/v1/traces"),
	)
}

func newTraceProvider(ctx context.Context, exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	r, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("ExampleService"),
		),
	)
	if err != nil {
		panic(err)
	}
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)

}

func main() {
	ctx := context.Background()
	exp, err := newExporter(ctx)
	if err != nil {
		log.Fatalf("error with newExporter: %v", err)
	}
	tp := newTraceProvider(ctx, exp)
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("failed to shutdown TraceProvider: %v", err)
		}
	}()

	otel.SetTracerProvider(tp)
	tracer = tp.Tracer("httpserver")

	metricsHandler := mymetrics.Handler()

	http1.HandleFunc("/metrics", func(w http1.ResponseWriter, r *http.Request) {
		metricsHandler.ServeHTTP(w, r)
	})

	http1.HandleFunc("/hello", func(w http1.ResponseWriter, r *http.Request) {
		_, span := tracer.Start(r.Context(), "HTTP GET /hello")
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", "/hello"),
		)

		fmt.Fprintf(w, "hello")
	})
	http1.HandleFunc("/echo", func(w http1.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		_, err := io.Copy(w, r.Body)
		if err != nil {
			//http.Error(w, "Failed to copy body", http.StatusInternalServerError)
			w.WriteHeader(500)
			io.WriteString(w, fmt.Sprintf("failed to copy body: %d", http.StatusInternalServerError))
		}

	})
	//store := NewUserStorage()
	//http.HandleFunc("GET /users", handleGetUsers(store))

	//tls
	/*if err := http1.ListenAndServeTLS(":8443", nil); err != nil {
		log.Fatalf("Server-Abbruch: %v", err)
	}*/

	//tcp

	if err := http1.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}

}
