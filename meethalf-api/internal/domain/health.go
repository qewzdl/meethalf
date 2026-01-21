package domain

const (
	HealthStatusOK   = "ok"
	HealthStatusFail = "fail"
)

type HealthStatus struct {
	Status       string             `json:"status"`
	Dependencies []HealthDependency `json:"dependencies,omitempty"`
}

type HealthDependency struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
	Timeout string `json:"timeout"`
}
