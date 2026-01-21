package health

import (
	"context"

	"meethalf-api/internal/domain"
)

type CompositeRepository struct {
	repos []Repository
}

func NewCompositeRepository(repos ...Repository) *CompositeRepository {
	filtered := make([]Repository, 0, len(repos))
	for _, repo := range repos {
		if repo != nil {
			filtered = append(filtered, repo)
		}
	}

	return &CompositeRepository{repos: filtered}
}

func (r *CompositeRepository) Check(ctx context.Context) []domain.HealthDependency {
	if ctx == nil {
		ctx = context.Background()
	}

	if r == nil || len(r.repos) == 0 {
		return []domain.HealthDependency{
			{
				Name:    "dependencies",
				Status:  domain.HealthStatusFail,
				Error:   "health repositories are not configured",
				Timeout: "0s",
			},
		}
	}

	deps := make([]domain.HealthDependency, 0, len(r.repos))
	for _, repo := range r.repos {
		if repo == nil {
			continue
		}

		timeout := repo.Timeout()
		dep := domain.HealthDependency{
			Name:    repo.Name(),
			Timeout: timeout.String(),
		}

		checkCtx := ctx
		var cancel context.CancelFunc
		if timeout > 0 {
			checkCtx, cancel = context.WithTimeout(ctx, timeout)
		}

		err := repo.Ping(checkCtx)
		if cancel != nil {
			cancel()
		}

		if err != nil {
			dep.Status = domain.HealthStatusFail
			dep.Error = err.Error()
		} else {
			dep.Status = domain.HealthStatusOK
		}

		deps = append(deps, dep)
	}

	if len(deps) == 0 {
		return []domain.HealthDependency{
			{
				Name:    "dependencies",
				Status:  domain.HealthStatusFail,
				Error:   "health repositories are not configured",
				Timeout: "0s",
			},
		}
	}

	return deps
}
