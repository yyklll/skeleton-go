package tracing

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/yyklll/skeleton/pkg/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type fakeTracer struct {
	name string
	log  []string
}

func (f *fakeTracer) NewFromString(parent, label string) (Span, bool) {
	panic("not implemented yet")
}

func (f *fakeTracer) New(parent Span, label string) Span {
	f.log = append(f.log, "span started")

	return &mockSpan{tracer: f}
}

func (f *fakeTracer) FromContext(ctx context.Context) (Span, bool) {
	return nil, false
}

func (f *fakeTracer) NewContext(parent context.Context, span Span) context.Context {
	return parent
}

func (f *fakeTracer) AddGrpcServerOptions(addInterceptors func(s grpc.StreamServerInterceptor, u grpc.UnaryServerInterceptor)) {
	panic("not implemented yet")
}

func (f *fakeTracer) AddGrpcClientOptions(addInterceptors func(s grpc.StreamClientInterceptor, u grpc.UnaryClientInterceptor)) {
	panic("not implemented yet")
}

func (f *fakeTracer) Close() error {
	panic("not implemented yet")
}

func (f *fakeTracer) assertNoSpanWith(t *testing.T, substr string) {
	t.Helper()
	for _, logLine := range f.log {
		if strings.Contains(logLine, substr) {
			t.Fatalf("expected to not find [%v] but found it in [%v]", substr, logLine)
		}
	}
}

type mockSpan struct {
	tracer *fakeTracer
}

func (m *mockSpan) Finish() {
	m.tracer.log = append(m.tracer.log, "span finished")
}

func (m *mockSpan) Annotate(key string, value interface{}) {
	m.tracer.log = append(m.tracer.log, fmt.Sprintf("key: %v values:%v", key, value))
}

func TestMain(m *testing.M) {
	log.InitGlobalLogger("debug")
	m.Run()
}

func TestFakeSpan(t *testing.T) {
	ctx := context.Background()

	// It should be safe to call all the usual methods as if a plugin were installed.
	span1, ctx := NewSpan(ctx, "label")
	span1.Finish()

	span2, ctx := NewSpan(ctx, "label")
	span2.Annotate("key", 42)
	span2.Finish()

	span3, _ := NewSpan(ctx, "label")
	span3.Annotate("key", 42)
	span3.Finish()
}

func TestRegisterService(t *testing.T) {
	fakeName := "test"
	tracingBackendFactories[fakeName] = func(serviceName string) (tracingService, io.Closer, bool) {
		tracer := &fakeTracer{name: serviceName}
		return tracer, tracer, true
	}

	tracingServer = &fakeName

	serviceName := "service"
	closer := StartTracing(serviceName)
	tracer, ok := closer.(*fakeTracer)
	if !ok {
		t.Fatalf("did not get the expected tracer")
	}

	if tracer.name != serviceName {
		t.Fatalf("expected the name to be `%v` but it was `%v`", serviceName, tracer.name)
	}
}
