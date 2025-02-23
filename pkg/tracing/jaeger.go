package tracing

import (
	"flag"
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"

	"github.com/yyklll/skeleton/pkg/log"
)

var (
	agentHost    = flag.String("jaeger-agent-host", "", "host and port to send spans to. if empty, no tracing will be done")
	samplingRate = flag.Float64("tracing-sampling-rate", 0.1, "sampling rate for the probabilistic jaeger sampler")
)

// newJagerTracerFromEnv will instantiate a tracingService implemented by Jaeger,
// taking configuration from environment variables. Available properties are:
// JAEGER_SERVICE_NAME -- If this is set, the service name used in code will be ignored and this value used instead
// JAEGER_RPC_METRICS
// JAEGER_TAGS
// JAEGER_SAMPLER_TYPE
// JAEGER_SAMPLER_PARAM
// JAEGER_SAMPLER_MANAGER_HOST_PORT
// JAEGER_SAMPLER_MAX_OPERATIONS
// JAEGER_SAMPLER_REFRESH_INTERVAL
// JAEGER_REPORTER_MAX_QUEUE_SIZE
// JAEGER_REPORTER_FLUSH_INTERVAL
// JAEGER_REPORTER_LOG_SPANS
// JAEGER_ENDPOINT
// JAEGER_USER
// JAEGER_PASSWORD
// JAEGER_AGENT_HOST
// JAEGER_AGENT_PORT
func newJagerTracerFromEnv(serviceName string) (tracingService, io.Closer, bool) {
	cfg, err := config.FromEnv()
	if err != nil {
		return nil, nil, false
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = serviceName
	}

	// Allow command line args to override environment variables.
	if *agentHost != "" {
		cfg.Reporter.LocalAgentHostPort = *agentHost
	}
	log.Infof("Tracing to: %v as %v", cfg.Reporter.LocalAgentHostPort, cfg.ServiceName)
	cfg.Sampler = &config.SamplerConfig{
		Type:  jaeger.SamplerTypeConst,
		Param: *samplingRate,
	}
	log.Infof("Tracing sampling rate: %v", *samplingRate)

	tracer, closer, err := cfg.NewTracer()

	if err != nil {
		return nil, &nilCloser{}, false
	}

	opentracing.SetGlobalTracer(tracer)

	return openTracingService{Tracer: &jaegerTracer{actual: tracer}}, closer, false
}

func init() {
	tracingBackendFactories["opentracing-jaeger"] = newJagerTracerFromEnv
}

var _ tracer = (*jaegerTracer)(nil)

type jaegerTracer struct {
	actual opentracing.Tracer
}

func (jt *jaegerTracer) GetOpenTracingTracer() opentracing.Tracer {
	return jt.actual
}
