package httpserver

import "meethalf-api/internal/usecase/health"

type Handlers struct {
	Health *HealthHandler
}

func NewHandlers(healthUC health.Usecase) *Handlers {
	return &Handlers{
		Health: NewHealthHandler(healthUC),
	}
}
