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
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"

	"github.com/naumov-andrey/exploring-cloud-native/echo-server/config"
	"github.com/naumov-andrey/exploring-cloud-native/echo-server/tracing"
)

func main() {
	log.Printf("Staring %s:%s", config.ServiceName, config.ServiceVersion)

	ctx := context.Background()

	tracerProvider, err := tracing.NewJaegerTraceProvider()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	otel.SetTracerProvider(tracerProvider)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(otelecho.Middleware(config.ServiceName))

	e.POST("/", echoHandler)

	errCh := make(chan error, 1)
	go func() {
		err := e.Start(config.Port)
		errCh <- err
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

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
	msg, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	return c.String(http.StatusOK, string(msg))
}
