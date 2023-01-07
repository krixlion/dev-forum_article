package rabbitmq

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	QueueSize         int           // Max number of messages internally queued for publishing.
	MaxWorkers        int           // Max number of concurrent workers per operation type.
	ReconnectInterval time.Duration // Time between reconnect attempts.

	// Settings for the internal circuit breaker.
	MaxRequests   uint32        // Number of requests allowed to half-open state.
	ClearInterval time.Duration // Time after which failed calls count is cleared.
	ClosedTimeout time.Duration // Time after which closed state becomes half-open.
}

func DefaultConfig() Config {
	return Config{
		QueueSize:         100,              // Max number of messages internally queued for publishing.
		MaxWorkers:        30,               // Max number of concurrent workers per operation type.
		ReconnectInterval: time.Second * 2,  // Time between reconnect attempts.
		MaxRequests:       10,               // Number of requests allowed to half-open state.
		ClearInterval:     time.Second * 10, // Time after which failed calls count is cleared.
		ClosedTimeout:     time.Second * 10, // Time after which closed state becomes half-open.
	}
}

type Logger interface {
	Log(ctx context.Context, msg string, keyvals ...interface{})
}

type nullLogger struct{}

func (nullLogger) Log(context.Context, string, ...any) {}

type nullTracer struct{}

func (nullTracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return ctx, nullSpan{}
}

type nullSpan struct{}

func (nullSpan) End(options ...trace.SpanEndOption)                  {}
func (nullSpan) AddEvent(name string, options ...trace.EventOption)  {}
func (nullSpan) IsRecording() bool                                   { return false }
func (nullSpan) RecordError(err error, options ...trace.EventOption) {}
func (nullSpan) SpanContext() trace.SpanContext                      { return trace.SpanContext{} }
func (nullSpan) SetStatus(code codes.Code, description string)       {}
func (nullSpan) SetName(name string)                                 {}
func (nullSpan) SetAttributes(kv ...attribute.KeyValue)              {}
func (nullSpan) TracerProvider() trace.TracerProvider                { return nil }

func setSpanErr(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}
