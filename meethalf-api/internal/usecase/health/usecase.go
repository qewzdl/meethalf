package health

import (
	"context"
	"time"

	"meethalf-api/internal/domain"
)

type Usecase interface {
	Liveness(ctx context.Context) domain.HealthStatus
	Readiness(ctx context.Context) domain.HealthStatus
	Status(ctx context.Context) domain.HealthStatus
}

type Repository interface {
	Name() string
	Timeout() time.Duration
	Ping(ctx context.Context) error
}

type service struct {
	checker Checker
}

type Checker interface {
	Check(ctx context.Context) []domain.HealthDependency
}

func New(checker Checker) Usecase {
	return &service{checker: checker}
}

func (s *service) Liveness(ctx context.Context) domain.HealthStatus {
	return domain.HealthStatus{Status: domain.HealthStatusOK}
}

func (s *service) Readiness(ctx context.Context) domain.HealthStatus {
	deps := s.dependencies(ctx)
	return domain.HealthStatus{Status: statusFromDependencies(deps)}
}

func (s *service) Status(ctx context.Context) domain.HealthStatus {
	deps := s.dependencies(ctx)
	return domain.HealthStatus{
		Status:       statusFromDependencies(deps),
		Dependencies: deps,
	}
}

func (s *service) dependencies(ctx context.Context) []domain.HealthDependency {
	if s == nil || s.checker == nil {
		return []domain.HealthDependency{
			{
				Name:    "dependencies",
				Status:  domain.HealthStatusFail,
				Error:   "health repository is nil",
				Timeout: "0s",
			},
		}
	}

	return s.checker.Check(ctx)
}

func statusFromDependencies(deps []domain.HealthDependency) string {
	if len(deps) == 0 {
		return domain.HealthStatusFail
	}

	for _, dep := range deps {
		if dep.Status != domain.HealthStatusOK {
			return domain.HealthStatusFail
		}
	}

	return domain.HealthStatusOK
}
