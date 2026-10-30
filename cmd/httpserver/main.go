package main

import (
	"context"
	"fmt"
	"httpServer/internal/http1"
	"httpServer/internal/mymetrics"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
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
	otel.SetTextMapPropagator(propagation.TraceContext{})
	tracer = tp.Tracer("httpserver")

	metricsHandler := mymetrics.Handler()

	http1.HandleFunc("/metrics", func(w http1.ResponseWriter, r *http.Request) {
		metricsHandler.ServeHTTP(w, r)
	})

	http1.HandleFunc("/hello", func(w http1.ResponseWriter, r *http.Request) {
		// 1. Parent-Span starten
		ctx, span := tracer.Start(r.Context(), "HTTP GET /hello")
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", "/hello"),
		)

		// 2. Request an /dice mit Trace-Kontext erstellen
		diceReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:8080/dice", nil)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to create request for dice")
			span.SetAttributes(semconv.HTTPResponseStatusCode(http.StatusInternalServerError))
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "internal server error\n")
			return
		}

		// 3. Trace-Kontext in die HTTP-Header injizieren (W3C traceparent)
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(diceReq.Header))

		// 4. /dice aufrufen
		client := &http.Client{}
		resp, err := client.Do(diceReq)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to call dice service")
			span.SetAttributes(semconv.HTTPResponseStatusCode(http.StatusInternalServerError))
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, fmt.Sprintf("failed to reach dice: %v\n", err))
			return
		}
		defer resp.Body.Close()

		// 5. Prüfen, ob /dice einen 500er geworfen hat
		if resp.StatusCode == http.StatusInternalServerError {
			diceBody, _ := io.ReadAll(resp.Body)
			errMsg := fmt.Sprintf("dice downstream error 500: %s", string(diceBody))

			// Span als Fehler kennzeichnen (roter Balken in Grafana)
			span.SetStatus(codes.Error, errMsg)
			span.SetAttributes(semconv.HTTPResponseStatusCode(http.StatusInternalServerError))

			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, fmt.Sprintf("hello failed because dice failed: %s", string(diceBody)))
			return
		}

		// 6. Erfolgsfall (200 OK)
		diceBody, _ := io.ReadAll(resp.Body)
		span.SetStatus(codes.Ok, "success")
		span.SetAttributes(semconv.HTTPResponseStatusCode(http.StatusOK))

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "hello -> dice response: %s", string(diceBody))
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
	http1.HandleFunc("/dice", func(w http1.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		extractedCtx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		_, span := tracer.Start(extractedCtx, "HTTP GET /dice")
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", "/dice"),
		)

		if rand.Intn(5) == 0 {
			span.SetStatus(codes.Error, "error in /dice")
			span.SetAttributes(semconv.HTTPResponseStatusCode(http.StatusInternalServerError))
			w.WriteHeader(500)
			io.WriteString(w, fmt.Sprintf("dice rolled: %d", http.StatusInternalServerError))
			return
		}

		span.SetAttributes(semconv.HTTPResponseStatusCode(http.StatusOK))
		span.SetStatus(codes.Ok, "success")

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, fmt.Sprintf("dice rolled: %d\n", http.StatusOK))
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
