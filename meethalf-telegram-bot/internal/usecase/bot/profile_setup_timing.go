package bot

import "meethalf-telegram-bot/internal/domain"

type profileSetupStepTiming struct {
	Step    domain.ProfileDraftStep
	Seconds int
}

var profileSetupTimings = []profileSetupStepTiming{
	{Step: domain.ProfileDraftStepName, Seconds: 10},
	{Step: domain.ProfileDraftStepGender, Seconds: 2},
	{Step: domain.ProfileDraftStepBirthDate, Seconds: 5},
	{Step: domain.ProfileDraftStepCountry, Seconds: 5},
	{Step: domain.ProfileDraftStepCity, Seconds: 10},
	{Step: domain.ProfileDraftStepDescription, Seconds: 40},
	{Step: domain.ProfileDraftStepEmoji, Seconds: 10},
	{Step: domain.ProfileDraftStepPhotos, Seconds: 40},
}

func (s *service) profileSetupTotalSeconds() int {
	total := 0
	for _, timing := range profileSetupTimings {
		total += timing.Seconds
	}
	return total
}

func (s *service) profileSetupTotalMinutes() int {
	totalSeconds := s.profileSetupTotalSeconds()
	minutes := (totalSeconds + 30) / 60
	if minutes < 1 {
		minutes = 1
	}
	return minutes
}
