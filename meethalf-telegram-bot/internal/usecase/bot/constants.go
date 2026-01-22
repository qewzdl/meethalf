package bot

const (
	minAge               = 1
	maxAge               = 120
	maxNameLength        = 64
	maxCityLength        = 64
	maxDescriptionLength = 500
	minPhotos            = 1
	maxPhotos            = 4
	profileStepsTotal    = 8
	birthDateLayout      = "2006-01-02"
)

type profileStatus int

const (
	profileStatusUnknown profileStatus = iota
	profileStatusMissing
	profileStatusPresent
)
