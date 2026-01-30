package httpserver

import (
	"meethalf-api/internal/usecase/health"
	"meethalf-api/internal/usecase/matching"
	"meethalf-api/internal/usecase/moderation"
	"meethalf-api/internal/usecase/profile"
)

type Handlers struct {
	Health  *HealthHandler
	Profile *ProfileHandler
	Search  *SearchHandler
	Admin   *AdminHandler
}

func NewHandlers(healthUC health.Usecase, profileUC profile.Usecase, matchingUC matching.Usecase, moderationUC moderation.Usecase) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(healthUC),
		Profile: NewProfileHandler(profileUC),
		Search:  NewSearchHandler(matchingUC),
		Admin:   NewAdminHandler(moderationUC, matchingUC),
	}
}
