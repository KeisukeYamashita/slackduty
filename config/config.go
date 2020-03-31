package config

import (
	"errors"
	"io/ioutil"
	"os"
	"os/user"

	"gopkg.in/yaml.v2"
)

var (
	defaultConfigPath = "/.slackduty/config.yml"
)

// Config is the CLI configuration kept in SLACKDUTY_CONFIG(default value is )
type Config struct {
	Groups []Group `yaml:"groups"`
}

// Group represents one single rule for syncronizing.
// A group will syncronize with the same fetch schedule.
type Group struct {
	Name       string   `yaml:"name"`
	Exclude    []string `yaml:"exclude"`
	Members    *Members `yaml:"members"`
	Schedule   string   `yaml:"schedule"`
	Usergroups []string `yaml:"usergroups"`
}

// Members represents the Slack or Pagerduty user which belongs
// to the handle(s) defined in the same group.
type Members struct {
	Slack     *Slack     `yaml:"slack"`
	Pagerduty *Pagerduty `yaml:"pagerduty"`
}

// Load loads the config.yml from the filepath given.
func Load(path string) (*Config, error) {
	cfg := &Config{}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(bytes, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// GetConfigPath retrieves the config.yml file
func GetConfigPath() string {
	if path := os.Getenv("SLACKDUTY_CONFIG"); path != "" {
		return path
	}

	userHome := ""
	usr, err := user.Current()
	if err != nil {
		// Fallback by reading $HOME environment variables
		userHome = os.Getenv("HOME")
	} else {
		userHome = usr.HomeDir
	}

	return userHome + defaultConfigPath
}

// GetAPIKeys retrieves the API key from environment variables
func GetAPIKeys() (pdAPIKey, slackAPIKey string, err error) {
	if pdAPIKey = os.Getenv("SLACKDUTY_PAGERDUTY_API_KEY"); pdAPIKey == "" {
		return "", "", errors.New("SLACKDUTY_PAGERDUTY_API_KEY is not configured")
	}

	if slackAPIKey = os.Getenv("SLACKDUTY_SLACK_API_KEY"); slackAPIKey == "" {
		return pdAPIKey, "", errors.New("SLACKDUTY_SLACK_API_KEY is not configured")
	}

	return pdAPIKey, slackAPIKey, nil
}

// IsExternalTrigger retrieves if the Slackduty is triggered from external trigger or not.
// If it is configured to `true`, the Slackduty process will exits once it updates the Slack
// usergroup. It will ignore the `group[].schedule`.
//
// If `false`, Slackduty will run kind of like a daemon and runs by schedule specified.
func IsExternalTrigger() bool {
	if v := os.Getenv("SLACKDUTY_EXTERNAL_TRIGGER"); v == "true" {
		return true
	}

	return false
}
