package observability

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	gracefulShutdownTimeout = 5 * time.Second
)

func InitOtel(ctx context.Context, connStr string, serviceName string) error {
	res, err := resource.New(
		ctx,
		resource.WithAttributes(semconv.ServiceName(serviceName)),
	)
	if err != nil {
		return fmt.Errorf("failed to set otel resource: %v", err)
	}

	conn, err := grpc.NewClient(
		connStr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to otel: %v", err)
	}

	tp, err := initTracerProvider(ctx, conn, res)
	if err != nil {
		return fmt.Errorf("failed to initiate tracer provider: %v", err)
	}
	otel.SetTracerProvider(tp)

	mp, err := initMetricProvider(ctx, conn, res)
	if err != nil {
		return fmt.Errorf("failed to initiate metric provider: %v", err)
	}
	otel.SetMeterProvider(mp)

	go func() {
		<-ctx.Done()
		log.Printf("context canceled, shutting down tracer and metric providers\n")

		ctxGrace, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer cancel()
		if err := tp.Shutdown(ctxGrace); err != nil {
			log.Printf("failed to shut down tracer provider: %v\n", err)
		}
		if err := mp.Shutdown(ctxGrace); err != nil {
			log.Printf("failed to shut down meter provider: %v\n", err)
		}
	}()

	return nil
}

func initTracerProvider(ctx context.Context, conn *grpc.ClientConn, res *resource.Resource) (*trace.TracerProvider, error) {
	tracerExp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to initiate tracer exporter: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(tracerExp),
		trace.WithResource(res),
	)

	return tp, nil
}

func initMetricProvider(ctx context.Context, conn *grpc.ClientConn, res *resource.Resource) (*metric.MeterProvider, error) {
	// TODO: Create a custom metrics template for gowrap
	// or use the prometheus gowrap template and prometheus exporter.
	metricsExp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to initiate metric exporter: %v", err)
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricsExp)),
		metric.WithResource(res),
	)

	return mp, nil
}
