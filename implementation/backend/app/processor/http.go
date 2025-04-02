package processor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/pkg/logger"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsServer struct {
	server *http.Server
}

func NewMetricsServer(config *config.Config) *MetricsServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.WorkerMetricsPort),
		Handler: mux,
	}

	return &MetricsServer{
		server: server,
	}
}

func (s *MetricsServer) Run(ctx context.Context) error {
	l := logger.FromCtx(ctx)

	l.Sugar().Infof("Starting metrics server on %s\n", s.server.Addr)
	if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		l.Sugar().Errorw("metrics server failed", err)
		return err
	}

	return nil
}

func (s *MetricsServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
