package httpserver

import (
	"meethalf-api/internal/usecase/health"
	"meethalf-api/internal/usecase/profile"
)

type Handlers struct {
	Health  *HealthHandler
	Profile *ProfileHandler
}

func NewHandlers(healthUC health.Usecase, profileUC profile.Usecase) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(healthUC),
		Profile: NewProfileHandler(profileUC),
	}
}
