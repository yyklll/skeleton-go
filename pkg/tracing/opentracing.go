package tracing

import (
	"strings"

	otgrpc "github.com/opentracing-contrib/go-grpc"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var _ Span = (*openTracingSpan)(nil)

type openTracingSpan struct {
	otSpan opentracing.Span
}

// Finish will mark a span as finished
func (js openTracingSpan) Finish() {
	js.otSpan.Finish()
}

// Annotate will add information to an existing span
func (js openTracingSpan) Annotate(key string, value interface{}) {
	js.otSpan.SetTag(key, value)
}

type tracer interface {
	GetOpenTracingTracer() opentracing.Tracer
}

type openTracingService struct {
	Tracer tracer
}

func (jf openTracingService) AddGrpcServerOptions(addInterceptors func(s grpc.StreamServerInterceptor, u grpc.UnaryServerInterceptor)) {
	ot := jf.Tracer.GetOpenTracingTracer()
	addInterceptors(otgrpc.OpenTracingStreamServerInterceptor(ot), otgrpc.OpenTracingServerInterceptor(ot))
}

func (jf openTracingService) AddGrpcClientOptions(addInterceptors func(s grpc.StreamClientInterceptor, u grpc.UnaryClientInterceptor)) {
	ot := jf.Tracer.GetOpenTracingTracer()
	addInterceptors(otgrpc.OpenTracingStreamClientInterceptor(ot), otgrpc.OpenTracingClientInterceptor(ot))
}

func (jf openTracingService) NewClientSpan(parent *openTracingSpan, serviceName, label string) Span {
	span := jf.New(parent, label)
	span.Annotate("peer.service", serviceName)
	return span
}

// New is part of an interface implementation
func (jf openTracingService) New(parent Span, label string) Span {
	var innerSpan opentracing.Span
	if parent == nil {
		innerSpan = jf.Tracer.GetOpenTracingTracer().StartSpan(label)
	} else {
		jaegerParent := parent.(openTracingSpan)
		span := jaegerParent.otSpan
		innerSpan = jf.Tracer.GetOpenTracingTracer().StartSpan(label, opentracing.ChildOf(span.Context()))
	}
	return &openTracingSpan{otSpan: innerSpan}
}

func extractMapFromString(in string) (opentracing.TextMapCarrier, bool) {
	m := make(opentracing.TextMapCarrier)
	items := strings.Split(in, ":")
	if len(items) < 2 {
		return nil, false
	}
	for _, v := range items {
		idx := strings.Index(v, "=")
		if idx < 1 {
			return nil, false
		}
		m[v[0:idx]] = v[idx+1:]
	}
	return m, true
}

func (jf openTracingService) NewFromString(parent, label string) (Span, bool) {
	carrier, err := extractMapFromString(parent)
	if !err {
		return nil, false
	}

	spanContext, error := jf.Tracer.GetOpenTracingTracer().Extract(opentracing.TextMap, carrier)
	if error != nil {
		return nil, false
	}
	innerSpan := jf.Tracer.GetOpenTracingTracer().StartSpan(label, opentracing.ChildOf(spanContext))
	return &openTracingSpan{otSpan: innerSpan}, false
}

// FromContext is part of an interface implementation
func (jf openTracingService) FromContext(ctx context.Context) (Span, bool) {
	innerSpan := opentracing.SpanFromContext(ctx)

	if innerSpan == nil {
		return nil, false
	}
	return &openTracingSpan{otSpan: innerSpan}, true
}

// NewContext is part of an interface implementation
func (jf openTracingService) NewContext(parent context.Context, s Span) context.Context {
	span, ok := s.(openTracingSpan)
	if !ok {
		return nil
	}
	return opentracing.ContextWithSpan(parent, span.otSpan)
}
