package models

type HealthResult struct {
	Healthy      bool               `json:"healthy"`
	Dependencies []HealthResultItem `json:"dependencies"`
}

type HealthResultItem struct {
	Name    string `json:"name"`
	Healthy bool   `json:"healthy"`
	Error   string `json:"error"`
}
