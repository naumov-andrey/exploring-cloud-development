package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

const (
	serviceName    = "echo-server"
	serviceVersion = "v0.1.0"

	tracerName     = "tracer"
	tracesFileName = "traces.txt"
)

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	ctx := context.Background()

	f, err := os.Create(tracesFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	exporter, err := newExporter(f)
	if err != nil {
		log.Fatal(err)
	}

	resource, err := newResource()
	if err != nil {
		log.Fatal(err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource),
	)
	defer func() {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	otel.SetTracerProvider(tracerProvider)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.POST("/", echoHandler)

	errCh := make(chan error, 1)
	go func() {
		log.Printf("Starting server")
		err := e.Start(":8080")
		errCh <- err
	}()

	select {
	case <-sigCh:
		log.Printf("Starting shutdown...")

		if err := e.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	case err := <-errCh:
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Exiting")
}

func echoHandler(c echo.Context) error {
	ctx := c.Request().Context()

	_, span := otel.Tracer(tracerName).Start(ctx, "handler")
	defer span.End()

	msgBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		errMsg := fmt.Sprintf("failed to read request body: %s", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, errMsg)
		return c.String(http.StatusInternalServerError, errMsg)
	}

	msg := string(msgBytes)

	span.SetAttributes(attribute.String("request.message", msg))

	return c.String(http.StatusOK, msg)
}

func newExporter(w io.Writer) (trace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(w),
		stdouttrace.WithPrettyPrint(),
	)
}

func newResource() (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			attribute.String("environment", "demo"),
		),
	)
}
