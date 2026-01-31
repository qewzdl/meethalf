package bot

const (
	minAge                      = 16
	maxAge                      = 120
	maxNameLength               = 64
	maxCityLength               = 64
	maxDescriptionLength        = 500
	maxAdTextLength             = 4096
	maxAdCaptionLength          = 1024
	minPhotos                   = 1
	maxPhotos                   = 4
	adminAdBroadcastPageSize    = 200
	adminAdBroadcastQueueSize   = 5
	botCheckMinOperand          = 2
	botCheckMaxOperand          = 9
	botCheckMaxAttempts         = 3
	botCheckOptionsCount        = 4
	botCheckOptionsSpread       = 6
	botCheckOptionsMinValue     = 1
	botCheckOptionsColumns      = 2
	profileStepsTotal           = 9
	birthDateLayout             = "02.01.2006"
	birthDateExample            = "24.12.2006"
	legacyBirthDateLayout       = "2006-01-02"
	profileVisibilityHideAction = "hide"
	profileVisibilityShowAction = "show"
)

type profileStatus int

const (
	profileStatusUnknown profileStatus = iota
	profileStatusMissing
	profileStatusPresent
)
