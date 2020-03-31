package config

// Pagerduty ...
type Pagerduty struct {
	Schedules []string `yaml:"schedules"`
	Services  []string `yaml:"services"`
	Teams     []string `yaml:"teams"`
	Users     []string `yaml:"users"`
}
