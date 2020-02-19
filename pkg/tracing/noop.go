package tracing

import (
	"io"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type noopTracingServer struct{}

func (noopTracingServer) New(Span, string) Span                                     { return NoopSpan{} }
func (noopTracingServer) NewClientSpan(parent Span, serviceName, label string) Span { return NoopSpan{} }
func (noopTracingServer) FromContext(context.Context) (Span, bool)                  { return nil, false }
func (noopTracingServer) NewFromString(parent, label string) (Span, bool)           { return NoopSpan{}, true }
func (noopTracingServer) NewContext(parent context.Context, _ Span) context.Context { return parent }
func (noopTracingServer) AddGrpcServerOptions(addInterceptors func(s grpc.StreamServerInterceptor, u grpc.UnaryServerInterceptor)) {
}
func (noopTracingServer) AddGrpcClientOptions(addInterceptors func(s grpc.StreamClientInterceptor, u grpc.UnaryClientInterceptor)) {
}

// NoopSpan implements Span with no-op methods.
type NoopSpan struct{}

func (NoopSpan) Finish()                      {}
func (NoopSpan) Annotate(string, interface{}) {}

func init() {
	tracingBackendFactories["noop"] = func(_ string) (tracingService, io.Closer, bool) {
		return noopTracingServer{}, &nilCloser{}, true
	}
}
