package scylla

import (
	"context"
	"github.com/gocql/gocql"
	errors2 "github.com/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"strings"
	"time"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/pkg/logger"
)

type Scylla struct {
	Session *gocql.Session
}

func NewScylla(config *config.Config) (*Scylla, error) {
	hosts := strings.Split(config.ScyllaHosts, ",")
	cluster := gocql.NewCluster(hosts...)

	session, err := cluster.CreateSession()

	if err != nil {
		return nil, err
	}

	return &Scylla{
		Session: session,
	}, nil
}

func (s *Scylla) IsHealthy(baseCtx context.Context) error {
	l := logger.FromCtx(baseCtx).With(zap.String("service", "scylla-health-check"))

	ctx, cancel := context.WithTimeout(baseCtx, 5*time.Second)
	defer cancel()

	// Use a simple, low-cost query to check health
	var now time.Time

	q := s.Session.Query("SELECT now() FROM system.local").WithContext(ctx)

	if err := q.Scan(&now); err != nil {
		l.Error("scylla health check failed", zap.Error(err))
		return errors2.Wrap(err, "scylla is not healthy")
	}

	return nil
}

func (s *Scylla) Stop() {
	s.Session.Close()
}

var Module = fx.Options(
	fx.Provide(fx.Annotate(NewScylla,
		fx.OnStop(func(s *Scylla) {
			s.Stop()
		}),
	)),
)
