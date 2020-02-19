package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var QueryTotalCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "skeleton",
		Subsystem: "server",
		Name:      "query_total",
		Help:      "Counter of queries.",
	}, []string{"method"})
