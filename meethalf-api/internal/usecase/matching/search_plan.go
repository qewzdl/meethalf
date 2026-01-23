package matching

type searchAttempt struct {
	Accuracy  int
	AgeWindow int
}

type searchPlanner struct {
	minAccuracy int
	maxAccuracy int
	ageWindows  []int
}

// defaultAgeWindows maps accuracy (index) to the age window in years.
var defaultAgeWindows = []int{9, 7, 5, 3, 2}

func newSearchPlanner() searchPlanner {
	return searchPlanner{
		minAccuracy: minAccuracy,
		maxAccuracy: maxAccuracy,
		ageWindows:  append([]int(nil), defaultAgeWindows...),
	}
}

func (p searchPlanner) Attempts(requested int) []searchAttempt {
	planner := p.withDefaults()
	minAccuracy := planner.minAccuracy
	maxAccuracy := planner.maxAccuracy

	if requested < minAccuracy {
		requested = minAccuracy
	}
	if requested > maxAccuracy {
		requested = maxAccuracy
	}

	attempts := make([]searchAttempt, 0, requested-minAccuracy+1)
	for accuracy := requested; accuracy >= minAccuracy; accuracy-- {
		attempts = append(attempts, searchAttempt{
			Accuracy:  accuracy,
			AgeWindow: planner.ageWindow(accuracy),
		})
	}

	return attempts
}

func (p searchPlanner) ageWindow(accuracy int) int {
	if accuracy < p.minAccuracy {
		accuracy = p.minAccuracy
	}
	if accuracy > p.maxAccuracy {
		accuracy = p.maxAccuracy
	}
	if accuracy < 0 || accuracy >= len(p.ageWindows) {
		return defaultAgeWindow
	}

	return p.ageWindows[accuracy]
}

func (p searchPlanner) withDefaults() searchPlanner {
	planner := p
	if planner.maxAccuracy < planner.minAccuracy || planner.minAccuracy == 0 && planner.maxAccuracy == 0 {
		planner.minAccuracy = minAccuracy
		planner.maxAccuracy = maxAccuracy
	}
	if len(planner.ageWindows) == 0 {
		planner.ageWindows = append([]int(nil), defaultAgeWindows...)
	}

	return planner
}
