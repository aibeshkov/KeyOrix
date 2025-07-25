package interceptors

import (
	"context"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
)

// Metrics holds gRPC server metrics
type Metrics struct {
	TotalRequests   int64
	SuccessRequests int64
	ErrorRequests   int64
	TotalDuration   int64 // in nanoseconds
}

var grpcMetrics = &Metrics{}

// MetricsInterceptor returns a unary server interceptor for collecting metrics
func MetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Increment total requests
		atomic.AddInt64(&grpcMetrics.TotalRequests, 1)

		// Call the handler
		resp, err := handler(ctx, req)

		// Record duration
		duration := time.Since(start)
		atomic.AddInt64(&grpcMetrics.TotalDuration, duration.Nanoseconds())

		// Record success/error
		if err != nil {
			atomic.AddInt64(&grpcMetrics.ErrorRequests, 1)
		} else {
			atomic.AddInt64(&grpcMetrics.SuccessRequests, 1)
		}

		return resp, err
	}
}

// GetGRPCMetrics returns current gRPC metrics
func GetGRPCMetrics() *Metrics {
	return &Metrics{
		TotalRequests:   atomic.LoadInt64(&grpcMetrics.TotalRequests),
		SuccessRequests: atomic.LoadInt64(&grpcMetrics.SuccessRequests),
		ErrorRequests:   atomic.LoadInt64(&grpcMetrics.ErrorRequests),
		TotalDuration:   atomic.LoadInt64(&grpcMetrics.TotalDuration),
	}
}

// GetAverageResponseTime returns the average response time in milliseconds
func (m *Metrics) GetAverageResponseTime() float64 {
	if m.TotalRequests == 0 {
		return 0
	}
	return float64(m.TotalDuration) / float64(m.TotalRequests) / 1e6 // Convert to milliseconds
}

// GetSuccessRate returns the success rate as a percentage
func (m *Metrics) GetSuccessRate() float64 {
	if m.TotalRequests == 0 {
		return 0
	}
	return float64(m.SuccessRequests) / float64(m.TotalRequests) * 100
}

// GetErrorRate returns the error rate as a percentage
func (m *Metrics) GetErrorRate() float64 {
	if m.TotalRequests == 0 {
		return 0
	}
	return float64(m.ErrorRequests) / float64(m.TotalRequests) * 100
}
