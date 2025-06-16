package processor

import (
	"context"
	"fmt"
	"github.com/platinummonkey/go-concurrency-limits/core"
	"github.com/platinummonkey/go-concurrency-limits/limit"
	limiter2 "github.com/platinummonkey/go-concurrency-limits/limiter"
	"github.com/platinummonkey/go-concurrency-limits/metric_registry/gometrics"
	"github.com/platinummonkey/go-concurrency-limits/strategy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rcrowley/go-metrics"
	"go.uber.org/zap"
	"time"
	go_metrics_prometheus "tugas-akhir/backend/pkg/go-metrics-prometheus"
	"tugas-akhir/backend/pkg/logger"
)

type LimiterLogger struct {
	l *zap.Logger
}

func (l *LimiterLogger) Debugf(msg string, params ...interface{}) {
	l.l.Info(fmt.Sprintf(msg, params...))
}

func (l *LimiterLogger) IsDebugEnabled() bool {
	return true
}

func NewLimiter(ctx context.Context) (core.Limiter, *go_metrics_prometheus.PrometheusConfig, error) {
	metricsRegistry := metrics.NewRegistry()

	prometheusClient := go_metrics_prometheus.NewPrometheusProvider(
		metricsRegistry,
		ProcessorNamespace,
		ProcessorSubsystem,
		prometheus.DefaultRegisterer,
		PollInterval,
	)

	limiterMetricsRegistry, err := gometrics.NewGoMetricsMetricRegistry(
		metricsRegistry,
		"",
		"",
		PollInterval,
	)

	if err != nil {
		return nil, nil, err
	}

	limitLogger := LimiterLogger{
		l: logger.FromCtx(ctx),
	}

	// Setup concurrency limits
	limitStrategy := strategy.NewSimpleStrategyWithMetricRegistry(StrategyLimit, limiterMetricsRegistry)

	fixedLimit := limit.NewFixedLimit(LimiterName, 5000, limiterMetricsRegistry)

	//gradient2Limit, err := limit.NewGradient2Limit(
	//	LimiterName,
	//	1000,
	//	ConcurrencyLimit,
	//	500,
	//	func(limit int) int {
	//		return int(math.Max(4, float64(limit)/4))
	//	},
	//	0.2,
	//	1200,
	//	&limitLogger,
	//	limiterMetricsRegistry,
	//)
	//
	//if err != nil {
	//	return nil, nil, err
	//}

	tracedLimit := limit.NewTracedLimit(fixedLimit, &limitLogger)

	defaultLimiter, err := limiter2.NewDefaultLimiter(
		tracedLimit,
		int64(100*time.Millisecond),
		int64(500*time.Millisecond),
		int64(100*time.Microsecond),
		100,
		limitStrategy,
		&limitLogger,
		limiterMetricsRegistry,
	)

	if err != nil {
		return nil, nil, err
	}

	limiter := limiter2.NewQueueBlockingLimiterFromConfig(defaultLimiter, limiter2.QueueLimiterConfig{
		Ordering:          limiter2.OrderingFIFO,
		MaxBacklogTimeout: 15 * time.Minute,
		MaxBacklogSize:    10000,
		MetricRegistry:    limiterMetricsRegistry,
	})

	return limiter, prometheusClient, nil
}
