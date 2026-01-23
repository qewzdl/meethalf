package bot

const (
	minAge                      = 1
	maxAge                      = 120
	maxNameLength               = 64
	maxCityLength               = 64
	maxDescriptionLength        = 500
	minPhotos                   = 1
	maxPhotos                   = 4
	botCheckMinOperand          = 2
	botCheckMaxOperand          = 9
	botCheckMaxAttempts         = 3
	botCheckOptionsCount        = 4
	botCheckOptionsSpread       = 6
	botCheckOptionsMinValue     = 1
	botCheckOptionsColumns      = 2
	profileStepsTotal           = 9
	birthDateLayout             = "2006-01-02"
	profileVisibilityHideAction = "hide"
	profileVisibilityShowAction = "show"
)

type profileStatus int

const (
	profileStatusUnknown profileStatus = iota
	profileStatusMissing
	profileStatusPresent
)
