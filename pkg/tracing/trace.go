package tracing

import (
	"flag"
	"io"
	"strings"

	"github.com/yyklll/skeleton/pkg/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Span represents a unit of work within a trace. After creating a Span with
// NewSpan(), call one of the Start methods to mark the beginning of the work
// represented by this Span. Call Finish() when that work is done to record the
// Span. A Span may be reused by calling Start again.
type Span interface {
	Finish()
	// Annotate records a key/value pair associated with a Span. It should be
	// called between Start and Finish.
	Annotate(key string, value interface{})
}

// NewSpan creates a new Span with the currently installed tracing plugin.
// If no tracing plugin is installed, it returns a fake Span that does nothing.
func NewSpan(inCtx context.Context, label string) (Span, context.Context) {
	parent, _ := currentTracer.FromContext(inCtx)
	span := currentTracer.New(parent, label)
	outCtx := currentTracer.NewContext(inCtx, span)

	return span, outCtx
}

// NewFromString creates a new Span with the currently installed tracing plugin, extracting the span context from
// the provided string.
func NewFromString(inCtx context.Context, parent, label string) (Span, context.Context, bool) {
	span, err := currentTracer.NewFromString(parent, label)
	if !err {
		return nil, nil, false
	}
	outCtx := currentTracer.NewContext(inCtx, span)
	return span, outCtx, true
}

// FromContext returns the Span from a Context if present. The bool return
// value indicates whether a Span was present in the Context.
func FromContext(ctx context.Context) (Span, bool) {
	return currentTracer.FromContext(ctx)
}

// NewContext returns a context based on parent with a new Span value.
func NewContext(parent context.Context, span Span) context.Context {
	return currentTracer.NewContext(parent, span)
}

// CopySpan creates a new context from parentCtx, with only the trace span
// copied over from spanCtx, if it has any. If not, parentCtx is returned.
func CopySpan(parentCtx, spanCtx context.Context) context.Context {
	if span, ok := FromContext(spanCtx); ok {
		return NewContext(parentCtx, span)
	}
	return parentCtx
}

// AddGrpcServerOptions adds GRPC interceptors that read the parent span from the grpc packets
func AddGrpcServerOptions(addInterceptors func(s grpc.StreamServerInterceptor, u grpc.UnaryServerInterceptor)) {
	currentTracer.AddGrpcServerOptions(addInterceptors)
}

// AddGrpcClientOptions adds GRPC interceptors that add parent information to outgoing grpc packets
func AddGrpcClientOptions(addInterceptors func(s grpc.StreamClientInterceptor, u grpc.UnaryClientInterceptor)) {
	currentTracer.AddGrpcClientOptions(addInterceptors)
}

// tracingService is an interface for creating spans or extracting them from Contexts.
type tracingService interface {
	// New creates a new span from an existing one, if provided. The parent can also be nil
	New(parent Span, label string) Span

	// NewFromString creates a new span and uses the provided string to reconstitute the parent span
	NewFromString(parent, label string) (Span, bool)

	// FromContext extracts a span from a context, making it possible to annotate the span with additional information
	FromContext(ctx context.Context) (Span, bool)

	// NewContext creates a new context containing the provided span
	NewContext(parent context.Context, span Span) context.Context

	// AddGrpcServerOptions allows a tracing system to add interceptors to grpc server traffic
	AddGrpcServerOptions(addInterceptors func(s grpc.StreamServerInterceptor, u grpc.UnaryServerInterceptor))

	// AddGrpcClientOptions allows a tracing system to add interceptors to grpc server traffic
	AddGrpcClientOptions(addInterceptors func(s grpc.StreamClientInterceptor, u grpc.UnaryClientInterceptor))
}

// TracerFactory creates a tracing service for the service provided. It's important to close the provided io.Closer
// object to make sure that all spans are sent to the backend before the process exits.
type TracerFactory func(serviceName string) (tracingService, io.Closer, bool)

// tracingBackendFactories should be added to by a plugin during init() to install itself
var tracingBackendFactories = make(map[string]TracerFactory)

var currentTracer tracingService = noopTracingServer{}

var (
	tracingServer = flag.String("tracer", "noop", "tracing service to use")
)

// StartTracing enables tracing for a named service
func StartTracing(serviceName string) io.Closer {
	factory, ok := tracingBackendFactories[*tracingServer]
	if !ok {
		return fail(serviceName)
	}

	tracer, closer, err := factory(serviceName)
	if !err {
		log.Errorf("failed to create a %s tracer", *tracingServer)
		return &nilCloser{}
	}

	currentTracer = tracer

	log.Infof("successfully started tracing with [%s]", *tracingServer)

	return closer
}

func fail(serviceName string) io.Closer {
	options := make([]string, len(tracingBackendFactories))
	for k := range tracingBackendFactories {
		options = append(options, k)
	}
	altStr := strings.Join(options, ", ")
	log.Errorf("no such [%s] tracing service found. alternatives are: %v", serviceName, altStr)
	return &nilCloser{}
}

type nilCloser struct {
}

func (c *nilCloser) Close() error { return nil }
