package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	instrumentationVersion = "0.1.0"
)

var tracer trace.Tracer

func one(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "one")
	defer span.End()

	time.Sleep(25 * time.Millisecond)

	two(ctx)
}

func two(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "two")
	defer span.End()

	log.Println("Doing some important work!")
	time.Sleep(100 * time.Millisecond)

	span.SetAttributes(attribute.String("ImportantAttributeKey", "ImportantAttributeValue"))

	//Example of an error
	err := three(ctx)

	span.SetStatus(codes.Error, "operationThatCouldFail failed")
	span.RecordError(err)

}

func three(ctx context.Context) error {
	_, span := tracer.Start(ctx, "three")
	defer span.End()

	return fmt.Errorf("Exmaple of an ERROR")
}

func InitTracerProvider(ctx context.Context, serviceName string) func(context.Context) {
	//Init connection to otel-collector
	conn, _ := grpc.NewClient("otel-collector:4317",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	traceExporter, _ := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))

	// Add a custom service name to the resource
	serviceNameAttr := semconv.ServiceNameKey.String(serviceName)
	resourceWithServiceName, _ := sdkresource.New(ctx, sdkresource.WithAttributes(serviceNameAttr))

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(resourceWithServiceName),
	)

	otel.SetTracerProvider(tracerProvider)

	return nil
}

func NewTracer(ctx context.Context, serviceName string) *trace.Tracer {

	_ = InitTracerProvider(ctx, serviceName)

	t := otel.GetTracerProvider().Tracer("go.opentelemetry.io/otel", trace.WithInstrumentationVersion(instrumentationVersion))

	return &t
}

func main() {
	ctx := context.Background()
	// Initialisiere Tracer
	t := NewTracer(ctx, "BeispielProgramm")
	// Setze Tracer
	tracer = *t
	log.Println("Wait for Spans to get created.")

	one(ctx)

	log.Println("Wait for sending Spans to the collector")
	time.Sleep(60 * time.Second)
	log.Println("Done")

}
