package nulls

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type NullLogger struct{}

func (NullLogger) Log(context.Context, string, ...any) {}

type NullTracer struct{}

func (NullTracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return ctx, NullSpan{}
}

type NullSpan struct{}

func (NullSpan) End(options ...trace.SpanEndOption)                  {}
func (NullSpan) AddEvent(name string, options ...trace.EventOption)  {}
func (NullSpan) IsRecording() bool                                   { return false }
func (NullSpan) RecordError(err error, options ...trace.EventOption) {}
func (NullSpan) SpanContext() trace.SpanContext                      { return trace.SpanContext{} }
func (NullSpan) SetStatus(code codes.Code, description string)       {}
func (NullSpan) SetName(name string)                                 {}
func (NullSpan) SetAttributes(kv ...attribute.KeyValue)              {}
func (NullSpan) TracerProvider() trace.TracerProvider                { return nil }
