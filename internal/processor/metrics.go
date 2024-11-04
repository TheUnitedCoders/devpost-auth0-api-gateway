package processor

import (
	"context"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/domain"
)

var (
	requestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "processor_request_count",
			Help: "The total number of requests to processor",
		},
		[]string{"service", "method", "http_method", "status_code"},
	)

	requestTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "processor_request_work_time",
			Help:    "Processor request work time",
			Buckets: []float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120},
		},
		[]string{"service", "method", "http_method", "status_code"},
	)
)

type metricsMiddleware struct {
	next Processor
}

func (mw *metricsMiddleware) Process(ctx context.Context, request *domain.ProcessRequest) *domain.ProviderProcessResponse {
	startedAt := time.Now()

	resp := mw.next.Process(ctx, request)
	labels := []string{request.Service, request.APIMethod, request.HTTPMethod.String(), strconv.Itoa(int(resp.StatusCode))}

	requestCount.WithLabelValues(labels...).Inc()
	requestTime.WithLabelValues(labels...).Observe(time.Since(startedAt).Seconds())

	return resp
}

// WithMetricsMiddleware wraps Processor to handle Prometheus metrics.
func WithMetricsMiddleware(next Processor) Processor {
	return &metricsMiddleware{next: next}
}
